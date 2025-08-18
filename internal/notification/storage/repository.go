package storage

import (
	"context"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

type NotificationRepository interface {
	InsertDeviceToken(ctx context.Context, token *entity.DeviceToken) error
	GetDeviceTokensByUserID(ctx context.Context, userID string) ([]*entity.DeviceToken, error)
}

type notificationRepository struct {
	db boil.ContextExecutor
}

func NewNotificationRepository(db boil.ContextExecutor) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) InsertDeviceToken(ctx context.Context, deviceToken *entity.DeviceToken) error {
	err := deviceToken.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("failed to insert device token: %w", err)
	}

	return nil
}

func (r *notificationRepository) GetDeviceTokensByUserID(ctx context.Context, userID string) ([]*entity.DeviceToken, error) {
	tokens, err := entity.DeviceTokens(
		qm.Where("user_id = ?", userID),
	).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}
