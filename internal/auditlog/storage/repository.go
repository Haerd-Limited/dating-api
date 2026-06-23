//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
package storage

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/auditlog/domain"
)

type Repository interface {
	Insert(ctx context.Context, entry domain.Entry) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Insert(ctx context.Context, entry domain.Entry) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_audit_log (occurred_at, actor_ip, token_fp, method, path, target_id, status_code)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entry.OccurredAt, entry.ActorIP, entry.TokenFP, entry.Method, entry.Path, entry.TargetID, entry.StatusCode)
	if err != nil {
		return fmt.Errorf("insert admin audit log: %w", err)
	}

	return nil
}
