package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/adminsession/domain"
)

type Repository interface {
	Insert(ctx context.Context, session domain.Session) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)
	GetByID(ctx context.Context, id string) (*domain.Session, error)
	Touch(ctx context.Context, id string, lastSeenAt, expiresAt time.Time) error
	DeleteByTokenHash(ctx context.Context, tokenHash string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Insert(ctx context.Context, session domain.Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_sessions (id, display_name, token_hash, api_key_fp, ip, created_at, last_seen_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, session.ID, session.DisplayName, session.TokenHash, session.APIKeyFP, session.IP,
		session.CreatedAt, session.LastSeenAt, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("insert admin session: %w", err)
	}

	return nil
}

func (r *repository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	row := r.db.QueryRowxContext(ctx, `
		SELECT id, display_name, token_hash, api_key_fp, ip, created_at, last_seen_at, expires_at
		FROM admin_sessions
		WHERE token_hash = $1
	`, tokenHash)

	return scanSession(row)
}

func (r *repository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	row := r.db.QueryRowxContext(ctx, `
		SELECT id, display_name, token_hash, api_key_fp, ip, created_at, last_seen_at, expires_at
		FROM admin_sessions
		WHERE id = $1
	`, id)

	return scanSession(row)
}

func (r *repository) Touch(ctx context.Context, id string, lastSeenAt, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE admin_sessions
		SET last_seen_at = $2, expires_at = $3
		WHERE id = $1
	`, id, lastSeenAt, expiresAt)
	if err != nil {
		return fmt.Errorf("touch admin session: %w", err)
	}

	return nil
}

func (r *repository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM admin_sessions WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete admin session: %w", err)
	}

	return nil
}

func scanSession(row *sqlx.Row) (*domain.Session, error) {
	var s domain.Session

	var ip sql.NullString

	err := row.Scan(&s.ID, &s.DisplayName, &s.TokenHash, &s.APIKeyFP, &ip,
		&s.CreatedAt, &s.LastSeenAt, &s.ExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("scan admin session: %w", err)
	}

	if ip.Valid {
		s.IP = &ip.String
	}

	return &s, nil
}
