package notification

import (
	"context"
	"fmt"

	"firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"go.uber.org/zap"
	"google.golang.org/api/option"

	"github.com/Haerd-Limited/dating-api/internal/notification/domain"
	"github.com/Haerd-Limited/dating-api/internal/notification/mappers"
	"github.com/Haerd-Limited/dating-api/internal/notification/storage"
)

type Service interface {
	RegisterDeviceToken(ctx context.Context, input domain.RegisterDeviceTokenInput) error
	SendPushNotification(ctx context.Context, input domain.SendNotificationInput) error
	SendFollowNotification(ctx context.Context, followerID, followedID string) error
}

type notificationService struct {
	logger           *zap.Logger
	notificationRepo storage.NotificationRepository
	fcmClient        *messaging.Client
}

func NewNotificationService(
	logger *zap.Logger,
	notificationRepo storage.NotificationRepository,
	credsJSON string,
) (Service, error) {
	// opt := option.WithCredentialsFile(serviceAccountPath) //old version using FirebaseServiceAccountPath
	opt := option.WithCredentialsJSON([]byte(credsJSON)) // use JSON directly

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app: %w", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase messaging client: %w", err)
	}

	return &notificationService{
		logger:           logger,
		notificationRepo: notificationRepo,
		fcmClient:        client,
	}, nil
}

func (s *notificationService) RegisterDeviceToken(ctx context.Context, input domain.RegisterDeviceTokenInput) error {
	s.logger.Info("Registering device token...", zap.Any("userID", input.UserID))

	var domainDeviceToken domain.DeviceToken
	domainDeviceToken.UserID = input.UserID
	domainDeviceToken.Token = input.Token

	entity := mappers.ToEntity(domainDeviceToken)

	return s.notificationRepo.InsertDeviceToken(ctx, entity)
}

func (s *notificationService) SendPushNotification(ctx context.Context, input domain.SendNotificationInput) error {
	s.logger.Info("Sending push notification...", zap.String("title", input.Title), zap.Any("userID", input.UserID))

	// 1. Get device tokens from DB
	tokens, err := s.notificationRepo.GetDeviceTokensByUserID(ctx, input.UserID)
	if err != nil {
		s.logger.Error("failed to get device tokens", zap.Error(err))
		return err
	}

	if len(tokens) == 0 {
		s.logger.Info("no device tokens found for user", zap.String("userID", input.UserID))
		return nil
	}

	// 2. Send messages individually
	var successCount, failureCount int

	for _, token := range tokens {
		msg := &messaging.Message{
			Token: token.Token,
			Notification: &messaging.Notification{
				Title: input.Title,
				Body:  input.Body,
			},
			Data: input.Data,
		}

		_, err := s.fcmClient.Send(ctx, msg)
		if err != nil {
			s.logger.Error("failed to send push notification", zap.Error(err), zap.String("token", token.Token))

			failureCount++

			continue
		}

		successCount++
	}

	s.logger.Info("push notifications sent", zap.Int("success", successCount), zap.Int("failure", failureCount))

	return nil
}

func (s *notificationService) SendFollowNotification(ctx context.Context, followerID, followedID string) error {
	title := "New Follower"
	body := "You have a new follower!"

	return s.SendPushNotification(ctx, domain.SendNotificationInput{
		UserID: followedID,
		Title:  title,
		Body:   body,
	})
}
