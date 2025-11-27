package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/utils"
)

type VerificationRepository interface {
	// Return S3 object keys (or presigned URLs) for the user's PRIVATE profile photos
	GetUserPrivatePhotoKeys(ctx context.Context, userID string) ([]string, error)
	CreateAttempt(ctx context.Context, a entity.VerificationAttempt) error
	MarkAttempt(ctx context.Context, upd entity.VerificationAttempt) error
	SetUserPhotoVerified(ctx context.Context, userID string, attemptID string) error
	GetVerificationAttemptByUserIDAndSessionID(ctx context.Context, userID string, sessionID string) (*entity.VerificationAttempt, error)
	CheckIfPendingAttemptsExist(ctx context.Context, userID string) (*entity.VerificationAttempt, error)
	InvalidatePhotoVerification(ctx context.Context, userID string, tx *sql.Tx) error
}

type verificationRepository struct {
	db *sqlx.DB
}

func NewVerificationRepository(db *sqlx.DB) VerificationRepository {
	return &verificationRepository{
		db: db,
	}
}

func (r *verificationRepository) InvalidatePhotoVerification(ctx context.Context, userID string, tx *sql.Tx) error {
	exec := r.executor(tx)

	uvs, err := entity.FindUserVerificationStatus(ctx, exec, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if uvs == nil {
		return nil
	} // nothing to do

	uvs.PhotoVerified = false
	uvs.PhotoVerifiedAt = null.Time{} // null
	uvs.UpdatedAt = time.Now().UTC()
	_, err = uvs.Update(ctx, exec, boil.Whitelist(
		entity.UserVerificationStatusColumns.PhotoVerified,
		entity.UserVerificationStatusColumns.PhotoVerifiedAt,
		entity.UserVerificationStatusColumns.UpdatedAt,
	))

	return err
}

func (r *verificationRepository) executor(tx *sql.Tx) boil.ContextExecutor {
	if tx != nil {
		return tx
	}

	return r.db
}

func (r *verificationRepository) CheckIfPendingAttemptsExist(ctx context.Context, userID string) (*entity.VerificationAttempt, error) {
	attempt, err := entity.VerificationAttempts(
		entity.VerificationAttemptWhere.UserID.EQ(userID),
		entity.VerificationAttemptWhere.Status.EQ(entity.VerificationStatusPending),
		qm.OrderBy(entity.VerificationAttemptColumns.CreatedAt+" DESC"),
		qm.Limit(1),
	).One(ctx, r.db)
	if err != nil {
		return nil, err
	}

	return attempt, nil
}

func (r *verificationRepository) GetVerificationAttemptByUserIDAndSessionID(ctx context.Context, userID string, sessionID string) (*entity.VerificationAttempt, error) {
	va, err := entity.VerificationAttempts(
		entity.VerificationAttemptWhere.UserID.EQ(userID),
		entity.VerificationAttemptWhere.SessionID.EQ(null.StringFrom(sessionID)),
		qm.OrderBy(entity.VerificationAttemptColumns.CreatedAt+" DESC"),
		qm.Limit(1),
	).One(ctx, r.db)
	if err != nil {
		return nil, err
	}

	return va, nil
}

func (r *verificationRepository) CreateAttempt(ctx context.Context, a entity.VerificationAttempt) error {
	now := time.Now().UTC()
	a.CreatedAt = now
	a.UpdatedAt = now

	err := a.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("failed to insert verification attempt: %w", err)
	}

	userVerificationStatus := entity.UserVerificationStatus{
		UserID:        a.UserID,
		LastAttemptID: null.StringFrom(a.ID),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err = userVerificationStatus.Upsert(
		ctx, r.db,
		true,
		[]string{entity.UserVerificationStatusColumns.UserID}, // conflict columns
		boil.Whitelist( // columns to update on conflict
			entity.UserVerificationStatusColumns.LastAttemptID,
			entity.UserVerificationStatusColumns.UpdatedAt,
		),
		boil.Infer())
	if err != nil {
		return fmt.Errorf("failed to insert user verification status: %w", err)
	}

	return nil
}

func (r *verificationRepository) MarkAttempt(ctx context.Context, upd entity.VerificationAttempt) error {
	var cols []string

	if upd.Status != "" {
		cols = append(cols, entity.VerificationAttemptColumns.Status)
	}

	if upd.SessionID.Valid { // if you ever want to update it
		cols = append(cols, entity.VerificationAttemptColumns.SessionID)
	}

	if upd.LivenessScore.Valid {
		cols = append(cols, entity.VerificationAttemptColumns.LivenessScore)
	}

	if upd.MatchScore.Valid {
		cols = append(cols, entity.VerificationAttemptColumns.MatchScore)
	}

	if upd.ReasonCodes.Valid {
		cols = append(cols, entity.VerificationAttemptColumns.ReasonCodes)
	}

	if upd.BestFrameS3Key.Valid {
		cols = append(cols, entity.VerificationAttemptColumns.BestFrameS3Key)
	}

	// Always bump updated_at
	upd.UpdatedAt = time.Now().UTC()

	cols = append(cols, entity.VerificationAttemptColumns.UpdatedAt)

	if len(cols) == 0 {
		return nil // nothing to update
	}

	_, err := upd.Update(ctx, r.db, boil.Whitelist(cols...))
	if err != nil {
		return fmt.Errorf("failed to update verification attempt: %w", err)
	}

	return nil
}

func (r *verificationRepository) SetUserPhotoVerified(ctx context.Context, userID string, attemptID string) error {
	now := time.Now().UTC()

	uvs := &entity.UserVerificationStatus{
		UserID:          userID,
		PhotoVerified:   true,
		PhotoVerifiedAt: null.TimeFrom(now),
		LastAttemptID:   null.StringFrom(attemptID),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Upsert on user_id
	err := uvs.Upsert(
		ctx,
		r.db,
		true, // doUpdateOnConflict
		[]string{entity.UserVerificationStatusColumns.UserID}, // conflict col
		boil.Whitelist( // update on conflict
			entity.UserVerificationStatusColumns.PhotoVerified,
			entity.UserVerificationStatusColumns.PhotoVerifiedAt,
			entity.UserVerificationStatusColumns.LastAttemptID,
			entity.UserVerificationStatusColumns.UpdatedAt,
		),
		boil.Whitelist( // insert columns
			entity.UserVerificationStatusColumns.UserID,
			entity.UserVerificationStatusColumns.PhotoVerified,
			entity.UserVerificationStatusColumns.PhotoVerifiedAt,
			entity.UserVerificationStatusColumns.LastAttemptID,
			entity.UserVerificationStatusColumns.CreatedAt,
			entity.UserVerificationStatusColumns.UpdatedAt,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to update user verification status: %w", err)
	}

	return nil
}

// todo: require frontend to return keys instead of urls for photos and voicenotes. or update service to parse for key and store in db
func (r *verificationRepository) GetUserPrivatePhotoKeys(ctx context.Context, userID string) ([]string, error) {
	photos, err := entity.Photos(entity.PhotoWhere.UserID.EQ(null.StringFrom(userID))).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	keys := make([]string, len(photos))

	for i, p := range photos {
		key, s3Err := utils.S3KeyFromURL(p.URL)
		if s3Err != nil {
			return nil, fmt.Errorf("failed to get S3 key from URL: %w", s3Err)
		}

		keys[i] = key
	}

	return keys, nil
}
