package storage

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

type Repository interface {
	CreateFeedback(ctx context.Context, feedback *entity.Feedback, tx *sql.Tx) error
	CreateFeedbackAttachments(ctx context.Context, attachments []*entity.FeedbackAttachment, tx *sql.Tx) error
	ListByUserID(ctx context.Context, userID string) ([]*entity.Feedback, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateFeedback(ctx context.Context, feedback *entity.Feedback, tx *sql.Tx) error {
	exec := r.executor(tx)
	return feedback.Insert(ctx, exec, boil.Infer())
}

func (r *repository) CreateFeedbackAttachments(ctx context.Context, attachments []*entity.FeedbackAttachment, tx *sql.Tx) error {
	exec := r.executor(tx)

	for _, attachment := range attachments {
		if err := attachment.Insert(ctx, exec, boil.Infer()); err != nil {
			return err
		}
	}

	return nil
}

func (r *repository) executor(tx *sql.Tx) boil.ContextExecutor {
	if tx != nil {
		return tx
	}

	return r.db
}

func (r *repository) ListByUserID(ctx context.Context, userID string) ([]*entity.Feedback, error) {
	return entity.Feedbacks(entity.FeedbackWhere.UserID.EQ(userID)).All(ctx, r.db)
}
