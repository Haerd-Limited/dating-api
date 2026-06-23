package profile

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/api/profile/dto"
	"github.com/Haerd-Limited/dating-api/internal/preference"
	internalprofile "github.com/Haerd-Limited/dating-api/internal/profile"
	"github.com/Haerd-Limited/dating-api/internal/profile/domain"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
)

type stubProfileService struct {
	getUserPromptTranscripts func(ctx context.Context, userID string) ([]domain.VoicePromptTranscript, error)
}

func (s *stubProfileService) GetUserPromptTranscripts(ctx context.Context, userID string) ([]domain.VoicePromptTranscript, error) {
	return s.getUserPromptTranscripts(ctx, userID)
}

func (s *stubProfileService) GetEnrichedProfile(context.Context, string) (domain.EnrichedProfile, error) {
	return domain.EnrichedProfile{}, nil
}

func (s *stubProfileService) GetProfileCard(context.Context, string) (profilecard.ProfileCard, error) {
	return profilecard.ProfileCard{}, nil
}

func (s *stubProfileService) GetProfileCardWithDistance(context.Context, string, float64, float64) (profilecard.ProfileCard, error) {
	return profilecard.ProfileCard{}, nil
}

func (s *stubProfileService) GetProfileForUpdate(context.Context, string) (domain.UpdateProfile, error) {
	return domain.UpdateProfile{}, nil
}

func (s *stubProfileService) UpdateProfile(context.Context, domain.UpdateProfile) error {
	return nil
}

func (s *stubProfileService) ScaffoldProfile(context.Context, *sql.Tx, string) error {
	return nil
}

func (s *stubProfileService) GetVoicePromptByID(context.Context, int64) (domain.VoicePrompt, error) {
	return domain.VoicePrompt{}, nil
}

func (s *stubProfileService) GetUserPhotos(context.Context, string) ([]domain.Photo, error) {
	return nil, nil
}

func (s *stubProfileService) GetTranscript(context.Context, int64) (string, error) {
	return "", nil
}

func (s *stubProfileService) UpsertUserSpokenLanguages(context.Context, string, []int16) error {
	return nil
}

func (s *stubProfileService) UpsertUserPhotos(context.Context, string, []domain.Photo) error {
	return nil
}

func (s *stubProfileService) UpsertUserPrompts(context.Context, string, []domain.VoicePromptUpdate) error {
	return nil
}

func (s *stubProfileService) VerifyProfile(context.Context, string) error {
	return nil
}

func (s *stubProfileService) IsVerified(context.Context, string) (bool, error) {
	return false, nil
}

func (s *stubProfileService) SetProfileUnderReview(context.Context, string) error {
	return nil
}

func (s *stubProfileService) SetProfileUnverified(context.Context, string) error {
	return nil
}

func (s *stubProfileService) CountBasicsCompletedByGender(context.Context, int16) (int64, error) {
	return 0, nil
}

func (s *stubProfileService) CountBasicsCompleted(context.Context) (int64, error) {
	return 0, nil
}

var _ internalprofile.Service = (*stubProfileService)(nil)

func TestSetAnalyticsOptOut(t *testing.T) {
	ctrl := gomock.NewController(t)
	prefSvc := preference.NewMockService(ctrl)
	h := NewProfileHandler(zaptest.NewLogger(t), nil, nil, nil, nil, prefSvc)

	cases := []struct {
		name     string
		optedOut bool
	}{
		{name: "opt out true", optedOut: true},
		{name: "opt in false", optedOut: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(dto.AnalyticsOptOutRequest{OptedOut: tc.optedOut})
			require.NoError(t, err)

			prefSvc.EXPECT().SetAnalyticsOptOut(gomock.Any(), "user-1", tc.optedOut).Return(nil)

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/preferences/analytics-opt-out", bytes.NewReader(body))
			req = req.WithContext(context.WithValue(req.Context(), commoncontext.UserIDKey, "user-1"))
			rec := httptest.NewRecorder()
			h.SetAnalyticsOptOut().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var resp dto.AnalyticsOptOutResponse

			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, tc.optedOut, resp.OptedOut)
		})
	}
}

func TestGetUserPromptTranscriptsSuccess(t *testing.T) {
	viewedUserID := "viewed-user-1"
	profileSvc := &stubProfileService{
		getUserPromptTranscripts: func(_ context.Context, userID string) ([]domain.VoicePromptTranscript, error) {
			assert.Equal(t, viewedUserID, userID)

			return []domain.VoicePromptTranscript{
				{PromptID: 1, Transcript: "hello"},
				{PromptID: 2, Transcript: "world"},
			}, nil
		},
	}
	h := NewProfileHandler(zaptest.NewLogger(t), profileSvc, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+viewedUserID+"/voice-prompts/transcripts", nil)
	req = req.WithContext(context.WithValue(req.Context(), commoncontext.UserIDKey, "viewer-1"))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", viewedUserID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	h.GetUserPromptTranscripts().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp dto.VoicePromptTranscriptsResponse

	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Transcripts, 2)
	assert.Equal(t, int64(1), resp.Transcripts[0].PromptID)
	assert.Equal(t, "hello", resp.Transcripts[0].Transcript)
}

func TestGetUserPromptTranscriptsNotFound(t *testing.T) {
	viewedUserID := "missing-user"
	profileSvc := &stubProfileService{
		getUserPromptTranscripts: func(_ context.Context, _ string) ([]domain.VoicePromptTranscript, error) {
			return nil, fmt.Errorf("get user voice prompts: %w", sql.ErrNoRows)
		},
	}
	h := NewProfileHandler(zaptest.NewLogger(t), profileSvc, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+viewedUserID+"/voice-prompts/transcripts", nil)
	req = req.WithContext(context.WithValue(req.Context(), commoncontext.UserIDKey, "viewer-1"))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", viewedUserID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	h.GetUserPromptTranscripts().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
