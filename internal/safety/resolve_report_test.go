package safety

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/auth"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	safetydomain "github.com/Haerd-Limited/dating-api/internal/safety/domain"
	safetystorage "github.com/Haerd-Limited/dating-api/internal/safety/storage"
	"github.com/Haerd-Limited/dating-api/internal/uow"
	"github.com/Haerd-Limited/dating-api/internal/user"
	userdomain "github.com/Haerd-Limited/dating-api/internal/user/domain"
)

const reportedUserID = "reported-1"

// fakeTx is a no-op uow.Tx whose Raw() returns a nil *sql.Tx. The mocked safety
// repository and user service ignore the tx, so no real database is touched.
type fakeTx struct {
	committed bool
}

func (t *fakeTx) Commit() error   { t.committed = true; return nil }
func (t *fakeTx) Rollback() error { return nil }
func (t *fakeTx) Raw() *sql.Tx    { return nil }

type fakeUoW struct {
	tx *fakeTx
}

func (u *fakeUoW) Begin(_ context.Context) (uow.Tx, error) { return u.tx, nil }

func newResolveTestService(
	t *testing.T,
	repo safetystorage.Repository,
	userSvc user.Service,
	authSvc auth.Service,
	tx *fakeTx,
) *service {
	t.Helper()

	return &service{
		logger:      zaptest.NewLogger(t),
		repo:        repo,
		uow:         &fakeUoW{tx: tx},
		userService: userSvc,
		authService: authSvc,
		// hub and notificationService intentionally nil: the post-commit
		// broadcast/push helpers short-circuit when they are nil.
	}
}

func baseResolveRequest(actionType string) safetydomain.ResolveReportRequest {
	notes := "policy violation"

	return safetydomain.ResolveReportRequest{
		ReportID:   "report-1",
		ReviewerID: "admin-1",
		ActionType: actionType,
		NewStatus:  safetydomain.ReportStatusResolved,
		Notes:      &notes,
	}
}

func expectReportLoadAndUpdate(repo *safetystorage.MockRepository) {
	repo.EXPECT().GetReportByID(gomock.Any(), "report-1").Return(&entity.UserReport{
		ID:             "report-1",
		ReportedUserID: reportedUserID,
	}, nil)
	repo.EXPECT().InsertReportAction(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	repo.EXPECT().UpdateReport(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
}

func TestResolveReport_WarnUser_InsertsWarning(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := safetystorage.NewMockRepository(ctrl)
	userSvc := user.NewMockService(ctrl)
	authSvc := auth.NewMockService(ctrl)
	tx := &fakeTx{}

	expectReportLoadAndUpdate(repo)

	repo.EXPECT().
		InsertModerationWarning(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, warning *entity.UserModerationWarning, _ *sql.Tx) error {
			assert.Equal(t, reportedUserID, warning.UserID)
			assert.Equal(t, "report-1", warning.ReportID.String)
			assert.Equal(t, "Please be respectful", warning.Message)
			warning.ID = "warning-1"

			return nil
		})

	// No account-status write and no session revocation for a warning.
	svc := newResolveTestService(t, repo, userSvc, authSvc, tx)

	req := baseResolveRequest(safetydomain.ActionWarnUser)
	req.ActionData = map[string]any{"warning_message": "Please be respectful"}

	require.NoError(t, svc.ResolveReport(context.Background(), req))
	assert.True(t, tx.committed)
}

func TestResolveReport_SuspendUser_SetsSuspendedAndRevokes(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := safetystorage.NewMockRepository(ctrl)
	userSvc := user.NewMockService(ctrl)
	authSvc := auth.NewMockService(ctrl)
	tx := &fakeTx{}

	expectReportLoadAndUpdate(repo)

	until := time.Now().UTC().Add(72 * time.Hour)

	userSvc.EXPECT().
		UpdateAccountStatus(gomock.Any(), reportedUserID, userdomain.AccountStatusSuspended, gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ string, suspendedUntil *time.Time, reason *string, _ *sql.Tx) error {
			require.NotNil(t, suspendedUntil)
			assert.True(t, suspendedUntil.After(time.Now().UTC()))
			require.NotNil(t, reason)

			return nil
		})

	authSvc.EXPECT().RevokeAllUserSessions(gomock.Any(), reportedUserID).Return(nil)

	svc := newResolveTestService(t, repo, userSvc, authSvc, tx)

	req := baseResolveRequest(safetydomain.ActionSuspendUser)
	req.ActionData = map[string]any{"suspend_until": until.Format(time.RFC3339)}

	require.NoError(t, svc.ResolveReport(context.Background(), req))
	assert.True(t, tx.committed)
}

func TestResolveReport_BanUser_SetsBannedAndRevokes(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := safetystorage.NewMockRepository(ctrl)
	userSvc := user.NewMockService(ctrl)
	authSvc := auth.NewMockService(ctrl)
	tx := &fakeTx{}

	expectReportLoadAndUpdate(repo)

	userSvc.EXPECT().
		UpdateAccountStatus(gomock.Any(), reportedUserID, userdomain.AccountStatusBanned, gomock.Nil(), gomock.Any(), gomock.Any()).
		Return(nil)

	authSvc.EXPECT().RevokeAllUserSessions(gomock.Any(), reportedUserID).Return(nil)

	svc := newResolveTestService(t, repo, userSvc, authSvc, tx)

	require.NoError(t, svc.ResolveReport(context.Background(), baseResolveRequest(safetydomain.ActionBanUser)))
	assert.True(t, tx.committed)
}

// TestResolveReport_NoUserEffect proves the gate is scoped: escalate / no_action
// commit the report workflow change but never touch the user's account or insert
// a warning. gomock fails the test if any unexpected method is called.
func TestResolveReport_NoUserEffect(t *testing.T) {
	for _, actionType := range []string{safetydomain.ActionEscalate, safetydomain.ActionNoAction} {
		t.Run(actionType, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := safetystorage.NewMockRepository(ctrl)
			userSvc := user.NewMockService(ctrl)
			authSvc := auth.NewMockService(ctrl)
			tx := &fakeTx{}

			expectReportLoadAndUpdate(repo)

			svc := newResolveTestService(t, repo, userSvc, authSvc, tx)

			require.NoError(t, svc.ResolveReport(context.Background(), baseResolveRequest(actionType)))
			assert.True(t, tx.committed)
		})
	}
}

func TestResolveReport_SuspendUser_MissingUntil_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := safetystorage.NewMockRepository(ctrl)
	userSvc := user.NewMockService(ctrl)
	authSvc := auth.NewMockService(ctrl)
	tx := &fakeTx{}

	expectReportLoadAndUpdate(repo)

	svc := newResolveTestService(t, repo, userSvc, authSvc, tx)

	err := svc.ResolveReport(context.Background(), baseResolveRequest(safetydomain.ActionSuspendUser))
	assert.ErrorIs(t, err, ErrInvalidSuspendUntil)
	assert.False(t, tx.committed)
}
