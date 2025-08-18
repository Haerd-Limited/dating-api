package storage

import (
	"context"
	"fmt"

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
}

type authRepository struct {
	db *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) AuthRepository {
	return &authRepository{
		db: db,
	}
}

func (r *authRepository) InsertRefreshToken(ctx context.Context, refreshToken *entity.RefreshToken) error {
	if err := refreshToken.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("failed inserting refresh token entity: %w", err)
	}

	return nil
}

func (r *authRepository) GetRefreshToken(ctx context.Context, refreshToken string) (*entity.RefreshToken, error) {
	rt, err := entity.RefreshTokens(entity.RefreshTokenWhere.Token.EQ(refreshToken)).One(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token by token %s: %w", refreshToken, err)
	}

	return rt, nil
}

func (r *authRepository) RevokeRefreshToken(ctx context.Context, refreshTokenID string) error {
	rt, err := entity.RefreshTokens(entity.RefreshTokenWhere.ID.EQ(refreshTokenID)).One(ctx, r.db)
	if err != nil {
		return fmt.Errorf("failed revoking refresh token entity: %w", err)
	}

	rt.Revoked = true

	_, err = rt.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("failed revoking refresh token entity: %w", err)
	}

	return nil
}

func (r *authRepository) RevokeAllRefreshTokens(ctx context.Context, userID string) error {
	_, err := entity.RefreshTokens(entity.RefreshTokenWhere.UserID.EQ(userID)).DeleteAll(ctx, r.db)
	if err != nil {
		return fmt.Errorf("failed revoking refresh token entity: %w", err)
	}

	return nil
}
