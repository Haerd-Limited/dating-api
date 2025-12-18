package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/verification/domain"
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
	// Video verification methods
	CheckIfPendingVideoAttemptExists(ctx context.Context, userID string) (*entity.VerificationAttempt, error)
	GetVideoAttemptByUserID(ctx context.Context, userID string) (*entity.VerificationAttempt, error)
	UpdateVideoAttemptWithKey(ctx context.Context, userID string, videoS3Key string) error
	GetVerificationCode(ctx context.Context, attemptID string) (string, error)
	UpdateVerificationCode(ctx context.Context, attemptID string, code string) error
	// Admin video verification methods
	ListVideoAttempts(ctx context.Context, filter domain.VideoAttemptFilter) ([]*entity.VerificationAttempt, error)
	GetVideoAttemptByID(ctx context.Context, attemptID string) (*entity.VerificationAttempt, error)
	UpdateVideoAttemptStatus(ctx context.Context, attemptID string, status string, rejectionReason *string) error
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

	if upd.VideoS3Key.Valid {
		cols = append(cols, entity.VerificationAttemptColumns.VideoS3Key)
	}

	if upd.VerificationCode.Valid {
		cols = append(cols, entity.VerificationAttemptColumns.VerificationCode)
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

func (r *verificationRepository) CheckIfPendingVideoAttemptExists(ctx context.Context, userID string) (*entity.VerificationAttempt, error) {
	attempt, err := entity.VerificationAttempts(
		entity.VerificationAttemptWhere.UserID.EQ(userID),
		entity.VerificationAttemptWhere.Type.EQ("video"),
		entity.VerificationAttemptWhere.Status.EQ(entity.VerificationStatusPending),
		qm.OrderBy(entity.VerificationAttemptColumns.CreatedAt+" DESC"),
		qm.Limit(1),
	).One(ctx, r.db)
	if err != nil {
		return nil, err
	}

	return attempt, nil
}

func (r *verificationRepository) GetVideoAttemptByUserID(ctx context.Context, userID string) (*entity.VerificationAttempt, error) {
	attempt, err := entity.VerificationAttempts(
		entity.VerificationAttemptWhere.UserID.EQ(userID),
		entity.VerificationAttemptWhere.Type.EQ("video"),
		qm.OrderBy(entity.VerificationAttemptColumns.CreatedAt+" DESC"),
		qm.Limit(1),
	).One(ctx, r.db)
	if err != nil {
		return nil, err
	}

	return attempt, nil
}

func (r *verificationRepository) UpdateVideoAttemptWithKey(ctx context.Context, userID string, videoS3Key string) error {
	attempt, err := r.GetVideoAttemptByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get video attempt: %w", err)
	}

	// Set video_s3_key and status fields
	// Note: After entity regeneration, these will be: attempt.VideoS3Key and attempt.VerificationCode
	// For now, we'll use reflection or update via raw SQL, but since MarkAttempt uses whitelist,
	// we need to handle the new fields. Let's use a direct update for now.
	attempt.Status = entity.VerificationStatusNeedsReview

	// Update via raw query since we need to set fields that don't exist in current entity
	_, err = r.db.ExecContext(ctx,
		`UPDATE verification_attempts SET video_s3_key = $1, status = $2, updated_at = now() WHERE id = $3`,
		videoS3Key,
		entity.VerificationStatusNeedsReview,
		attempt.ID,
	)
	if err != nil {
		return fmt.Errorf("update video attempt with key: %w", err)
	}

	return nil
}

func (r *verificationRepository) GetVerificationCode(ctx context.Context, attemptID string) (string, error) {
	var code sql.NullString

	err := r.db.QueryRowContext(ctx,
		"SELECT verification_code FROM verification_attempts WHERE id = $1", attemptID).Scan(&code)
	if err != nil {
		return "", fmt.Errorf("get verification code: %w", err)
	}

	if !code.Valid {
		return "", nil
	}

	return code.String, nil
}

func (r *verificationRepository) UpdateVerificationCode(ctx context.Context, attemptID string, code string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE verification_attempts SET verification_code = $1 WHERE id = $2", code, attemptID)
	if err != nil {
		return fmt.Errorf("update verification code: %w", err)
	}

	return nil
}

func (r *verificationRepository) ListVideoAttempts(ctx context.Context, filter domain.VideoAttemptFilter) ([]*entity.VerificationAttempt, error) {
	mods := []qm.QueryMod{
		entity.VerificationAttemptWhere.Type.EQ("video"),
		qm.OrderBy(entity.VerificationAttemptColumns.CreatedAt + " DESC"),
	}

	if len(filter.Status) > 0 {
		args := make([]interface{}, len(filter.Status))
		for i, status := range filter.Status {
			args[i] = status
		}

		mods = append(mods, qm.WhereIn(
			"verification_attempts.status IN ?",
			args...,
		))
	}

	if filter.UserID != nil && *filter.UserID != "" {
		mods = append(mods, entity.VerificationAttemptWhere.UserID.EQ(*filter.UserID))
	}

	if filter.Limit > 0 {
		mods = append(mods, qm.Limit(filter.Limit))
	} else {
		mods = append(mods, qm.Limit(50)) // default limit
	}

	if filter.Offset > 0 {
		mods = append(mods, qm.Offset(filter.Offset))
	}

	attempts, err := entity.VerificationAttempts(mods...).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("list video attempts: %w", err)
	}

	return attempts, nil
}

func (r *verificationRepository) GetVideoAttemptByID(ctx context.Context, attemptID string) (*entity.VerificationAttempt, error) {
	attempt, err := entity.VerificationAttempts(
		entity.VerificationAttemptWhere.ID.EQ(attemptID),
		entity.VerificationAttemptWhere.Type.EQ("video"),
	).One(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("get video attempt by id: %w", err)
	}

	return attempt, nil
}

func (r *verificationRepository) UpdateVideoAttemptStatus(ctx context.Context, attemptID string, status string, rejectionReason *string) error {
	attempt, err := entity.FindVerificationAttempt(ctx, r.db, attemptID)
	if err != nil {
		return fmt.Errorf("get video attempt: %w", err)
	}

	attempt.Status = status
	attempt.UpdatedAt = time.Now().UTC()

	// Store rejection reason in reason_codes JSON field, or clear it if approving
	if rejectionReason != nil {
		reasonCodesJSON, marshalErr := json.Marshal([]string{*rejectionReason})
		if marshalErr != nil {
			return fmt.Errorf("marshal rejection reason: %w", marshalErr)
		}

		attempt.ReasonCodes = null.JSONFrom(reasonCodesJSON)
	} else {
		// Clear reason codes when approving
		attempt.ReasonCodes = null.JSON{}
	}

	_, err = attempt.Update(ctx, r.db, boil.Whitelist(
		entity.VerificationAttemptColumns.Status,
		entity.VerificationAttemptColumns.ReasonCodes,
		entity.VerificationAttemptColumns.UpdatedAt,
	))
	if err != nil {
		return fmt.Errorf("update video attempt status: %w", err)
	}

	return nil
}
