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
	ListEvents(ctx context.Context, filter domain.ListFilter) ([]domain.Entry, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Insert(ctx context.Context, entry domain.Entry) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_audit_log (occurred_at, actor_ip, token_fp, method, path, target_id, status_code, actor_session_id, actor_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, entry.OccurredAt, entry.ActorIP, entry.TokenFP, entry.Method, entry.Path, entry.TargetID, entry.StatusCode, entry.ActorSessionID, entry.ActorName)
	if err != nil {
		return fmt.Errorf("insert admin audit log: %w", err)
	}

	return nil
}

func (r *repository) ListEvents(ctx context.Context, filter domain.ListFilter) ([]domain.Entry, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT occurred_at, actor_ip, token_fp, method, path, target_id, status_code, actor_session_id, actor_name
		FROM admin_audit_log
		WHERE method != 'GET' AND method != 'OPTIONS'
		  AND (
		    path LIKE '%/approve' OR path LIKE '%/reject' OR path LIKE '%/resolve'
		    OR path LIKE '%/broadcast' OR path LIKE '%/session'
		  )
	`
	args := []any{}
	argN := 1

	if filter.ActorName != nil && *filter.ActorName != "" {
		query += fmt.Sprintf(" AND actor_name = $%d", argN)

		args = append(args, *filter.ActorName)
		argN++
	}

	if filter.Action != nil && *filter.Action != "" {
		if pattern, ok := domain.ActionPathPattern(*filter.Action); ok {
			query += fmt.Sprintf(" AND path LIKE $%d", argN)

			args = append(args, pattern)
			argN++
		}
	}

	query += fmt.Sprintf(" ORDER BY occurred_at DESC LIMIT $%d OFFSET $%d", argN, argN+1)

	args = append(args, limit, filter.Offset)

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list admin audit events: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var out []domain.Entry

	for rows.Next() {
		var e domain.Entry
		if err := rows.Scan(&e.OccurredAt, &e.ActorIP, &e.TokenFP, &e.Method, &e.Path, &e.TargetID, &e.StatusCode, &e.ActorSessionID, &e.ActorName); err != nil {
			return nil, fmt.Errorf("scan admin audit event: %w", err)
		}

		out = append(out, e)
	}

	return out, rows.Err()
}
