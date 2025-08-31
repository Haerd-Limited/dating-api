package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
type AuthRepository interface {
	InsertRefreshToken(ctx context.Context, refreshToken *entity.RefreshToken) error
	GetRefreshToken(ctx context.Context, refreshToken string) (*entity.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, refreshTokenID string) error
	RevokeAllRefreshTokens(ctx context.Context, userID string) error
	CountRecentSends(ctx context.Context, channel, identifier, purpose string, since time.Time) (int, error)
	CountRecentSendsByIP(ctx context.Context, ip string, since time.Time) (int, error)
	InsertVerificationCode(ctx context.Context, vc *entity.VerificationCode) error
}

type authRepository struct {
	db *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) AuthRepository {
	return &authRepository{
		db: db,
	}
}

var ErrRefreshTokenNotFound = errors.New("refresh token not found")

func (r *authRepository) CountRecentSends(ctx context.Context, channel, identifier, purpose string, since time.Time) (int, error) {
	count, err := entity.VerificationCodes(
		entity.VerificationCodeWhere.Channel.EQ(channel),
		entity.VerificationCodeWhere.Identifier.EQ(identifier),
		entity.VerificationCodeWhere.Purpose.EQ(purpose),
		entity.VerificationCodeWhere.CreatedAt.GTE(since),
	).Count(ctx, r.db)
	if err != nil {
		return 0, err
	}
	return int(count), err
}

func (r *authRepository) CountRecentSendsByIP(ctx context.Context, ip string, since time.Time) (int, error) {
	count, err := entity.VerificationCodes(
		qm.Where("request_ip = ?::inet AND created_at >= ?", ip, since),
	).Count(ctx, r.db)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *authRepository) InsertVerificationCode(ctx context.Context, vc *entity.VerificationCode) error {
	err := vc.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return err
	}
	return nil
}

func (r *authRepository) InsertRefreshToken(ctx context.Context, refreshToken *entity.RefreshToken) error {
	if err := refreshToken.Insert(ctx, r.db, boil.Infer()); err != nil {
		return err
	}

	return nil
}

func (r *authRepository) GetRefreshToken(ctx context.Context, refreshToken string) (*entity.RefreshToken, error) {
	rt, err := entity.RefreshTokens(entity.RefreshTokenWhere.Token.EQ(refreshToken)).One(ctx, r.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}

		return nil, err
	}

	return rt, nil
}

func (r *authRepository) RevokeRefreshToken(ctx context.Context, refreshTokenID string) error {
	rt, err := entity.RefreshTokens(entity.RefreshTokenWhere.ID.EQ(refreshTokenID)).One(ctx, r.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRefreshTokenNotFound
		}

		return fmt.Errorf("repo auth select refresh token id=%s: %w", refreshTokenID, err)
	}

	rt.Revoked = true

	rows, err := rt.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("repo auth update refresh token id=%s: %w", refreshTokenID, err)
	}

	if rows == 0 { // deleted between read & write, or WHERE didn’t match
		return fmt.Errorf("repo auth update refresh token id=%s: %w", refreshTokenID, ErrRefreshTokenNotFound)
	}

	return nil
}

func (r *authRepository) RevokeAllRefreshTokens(ctx context.Context, userID string) error {
	_, err := entity.RefreshTokens(entity.RefreshTokenWhere.UserID.EQ(userID)).DeleteAll(ctx, r.db)
	if err != nil {
		return fmt.Errorf("repo auth delete refresh tokens userID=%s: %w", userID, err)
	}

	return nil
}
