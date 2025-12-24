package notification

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/api/option"

	"github.com/Haerd-Limited/dating-api/internal/notification/domain"
	"github.com/Haerd-Limited/dating-api/internal/notification/storage"
)

const (
	weeklyRefreshHour   = 0
	weeklyRefreshMinute = 5
	weeklyRefreshSecond = 0
	weeklyBatchSize     = 200
)

type Service interface {
	RegisterDeviceToken(ctx context.Context, userID, token string) error
	RemoveDeviceToken(ctx context.Context, userID, token string) error
	SendLikeNotification(ctx context.Context, likerID, likerName, recipientUserID string) error
	SendMatchNotification(ctx context.Context, counterpartName, recipientUserID, conversationID string) error
	SendNewMessageNotification(ctx context.Context, senderID, senderName, conversationID, recipientUserID, preview string) error
	SendVerificationApprovedNotification(ctx context.Context, recipientUserID string) error
	SendVerificationRejectedNotification(ctx context.Context, recipientUserID, rejectionReason string) error
	SendWeeklyRefreshNotifications(ctx context.Context) error
	StartWeeklyRefreshScheduler(ctx context.Context)
}

type service struct {
	logger            *zap.Logger
	deviceTokenRepo   storage.DeviceTokenRepository
	messaging         *messaging.Client
	weeklyJobStarted  atomic.Bool
	disableMessaging  bool
	credentialsSource string
}

type Config struct {
	ServiceAccountPath string
	CredentialsJSON    string
	ProjectID          string
}

func NewService(ctx context.Context, logger *zap.Logger, repo storage.DeviceTokenRepository, cfg Config) (Service, error) {
	messagingClient, credentialsSource, err := newMessagingClient(ctx, cfg)
	if err != nil {
		// If credentials are not provided, degrade gracefully but keep storing tokens.
		if errors.Is(err, errNoFirebaseCredentials) {
			logger.Sugar().Warn("firebase credentials not provided; push notifications disabled but device token APIs will continue to work")

			return &service{
				logger:            logger,
				deviceTokenRepo:   repo,
				messaging:         nil,
				disableMessaging:  true,
				credentialsSource: credentialsSource,
			}, nil
		}

		return nil, fmt.Errorf("initialise firebase messaging client: %w", err)
	}

	logger.Sugar().Infow("firebase messaging client initialised", "credentials_source", credentialsSource)

	return &service{
		logger:            logger,
		deviceTokenRepo:   repo,
		messaging:         messagingClient,
		credentialsSource: credentialsSource,
	}, nil
}

func (s *service) RegisterDeviceToken(ctx context.Context, userID, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("device token must not be empty")
	}

	if err := s.deviceTokenRepo.UpsertToken(ctx, userID, token); err != nil {
		return fmt.Errorf("store device token userID=%s: %w", userID, err)
	}

	return nil
}

func (s *service) RemoveDeviceToken(ctx context.Context, userID, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("device token must not be empty")
	}

	err := s.deviceTokenRepo.DeleteToken(ctx, userID, token)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("delete device token userID=%s: %w", userID, err)
	}

	return nil
}

func (s *service) SendLikeNotification(ctx context.Context, likerID, likerName, recipientUserID string) error {
	msg := domain.Message{
		Title: "New like",
		Body:  buildLikeBody(likerName),
		Data: map[string]string{
			"type":            "like.new",
			"liker_id":        likerID,
			"liker_name":      likerName,
			"recipient_id":    recipientUserID,
			"timestamp_utc":   time.Now().UTC().Format(time.RFC3339),
			"notification_id": "like.new",
		},
	}

	return s.sendToUsers(ctx, []string{recipientUserID}, msg)
}

func (s *service) SendMatchNotification(ctx context.Context, counterpartName, recipientUserID, conversationID string) error {
	msg := domain.Message{
		Title: "It's a match!",
		Body:  buildMatchBody(counterpartName),
		Data: map[string]string{
			"type":             "match.new",
			"conversation_id":  conversationID,
			"counterpart_name": counterpartName,
			"recipient_id":     recipientUserID,
			"timestamp_utc":    time.Now().UTC().Format(time.RFC3339),
			"notification_id":  "match.new",
		},
	}

	return s.sendToUsers(ctx, []string{recipientUserID}, msg)
}

func (s *service) SendNewMessageNotification(ctx context.Context, senderID, senderName, conversationID, recipientUserID, preview string) error {
	msg := domain.Message{
		Title: buildMessageTitle(senderName),
		Body:  preview,
		Data: map[string]string{
			"type":            "message.new",
			"conversation_id": conversationID,
			"sender_id":       senderID,
			"sender_name":     senderName,
			"recipient_id":    recipientUserID,
			"timestamp_utc":   time.Now().UTC().Format(time.RFC3339),
			"notification_id": "message.new",
		},
	}

	return s.sendToUsers(ctx, []string{recipientUserID}, msg)
}

func (s *service) SendVerificationApprovedNotification(ctx context.Context, recipientUserID string) error {
	msg := domain.Message{
		Title: "Verification approved",
		Body:  "Your video verification has been approved! Your profile is now verified.",
		Data: map[string]string{
			"type":            "verification.approved",
			"recipient_id":    recipientUserID,
			"timestamp_utc":   time.Now().UTC().Format(time.RFC3339),
			"notification_id": "verification.approved",
		},
	}

	return s.sendToUsers(ctx, []string{recipientUserID}, msg)
}

func (s *service) SendVerificationRejectedNotification(ctx context.Context, recipientUserID, rejectionReason string) error {
	body := "Your video verification has been rejected."
	if rejectionReason != "" {
		body = fmt.Sprintf("Your video verification has been rejected: %s", rejectionReason)
	}

	msg := domain.Message{
		Title: "Verification rejected",
		Body:  body,
		Data: map[string]string{
			"type":             "verification.rejected",
			"recipient_id":     recipientUserID,
			"rejection_reason": rejectionReason,
			"timestamp_utc":    time.Now().UTC().Format(time.RFC3339),
			"notification_id":  "verification.rejected",
		},
	}

	return s.sendToUsers(ctx, []string{recipientUserID}, msg)
}

func (s *service) SendWeeklyRefreshNotifications(ctx context.Context) error {
	userIDs, err := s.deviceTokenRepo.ListUserIDsWithTokens(ctx)
	if err != nil {
		return fmt.Errorf("list users with device tokens: %w", err)
	}

	if len(userIDs) == 0 {
		s.logger.Debug("weekly refresh: no users with device tokens")
		return nil
	}

	msg := domain.Message{
		Title: "Weekly refresh",
		Body:  "Your superlike and Voices Worth Hearing picks just refreshed. Come meet someone new.",
		Data: map[string]string{
			"type":            "weekly.refresh",
			"timestamp_utc":   time.Now().UTC().Format(time.RFC3339),
			"notification_id": "weekly.refresh",
		},
	}

	var sendErr error

	for start := 0; start < len(userIDs); start += weeklyBatchSize {
		end := start + weeklyBatchSize
		if end > len(userIDs) {
			end = len(userIDs)
		}

		chunk := userIDs[start:end]
		if err := s.sendToUsers(ctx, chunk, msg); err != nil {
			sendErr = multierr.Append(sendErr, err)
		}
	}

	return sendErr
}

func (s *service) StartWeeklyRefreshScheduler(ctx context.Context) {
	if s.disableMessaging {
		s.logger.Debug("weekly refresh scheduler not started (messaging disabled)")
		return
	}

	if !s.weeklyJobStarted.CompareAndSwap(false, true) {
		return
	}

	go s.runWeeklyRefreshScheduler(ctx)
}

func (s *service) runWeeklyRefreshScheduler(ctx context.Context) {
	timer := time.NewTimer(durationUntilNextSunday())
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Debug("weekly refresh scheduler stopping: context cancelled")
			return
		case <-timer.C:
			runCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			if err := s.SendWeeklyRefreshNotifications(runCtx); err != nil {
				s.logger.Sugar().Errorw("weekly refresh notification send failed", "error", err)
			} else {
				s.logger.Debug("weekly refresh notifications sent successfully")
			}

			cancel()

			timer.Reset(durationUntilNextSunday())
		}
	}
}

func (s *service) sendToUsers(ctx context.Context, userIDs []string, msg domain.Message) error {
	tokensByUser, err := s.deviceTokenRepo.GetTokensByUserIDs(ctx, userIDs)
	if err != nil {
		return fmt.Errorf("fetch device tokens: %w", err)
	}

	if len(tokensByUser) == 0 {
		return nil
	}

	var sendErr error

	for userID, tokens := range tokensByUser {
		if len(tokens) == 0 {
			continue
		}

		if err := s.sendToTokens(ctx, tokens, msg); err != nil {
			sendErr = multierr.Append(sendErr, fmt.Errorf("send push userID=%s: %w", userID, err))
		}
	}

	return sendErr
}

func (s *service) sendToTokens(ctx context.Context, tokens []string, msg domain.Message) error {
	if s.messaging == nil {
		s.logger.Debug("push messaging disabled; skipping send")
		return nil
	}

	messages := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: msg.Title,
			Body:  msg.Body,
		},
		Data: msg.Data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	responses, err := s.messaging.SendEachForMulticast(ctx, messages)
	if err != nil {
		return fmt.Errorf("send multicast: %w", err)
	}

	var invalidTokens []string

	for idx, res := range responses.Responses {
		if res == nil || res.Success {
			continue
		}

		if res.Error != nil && idx < len(tokens) && shouldRemoveToken(res.Error) {
			invalidTokens = append(invalidTokens, tokens[idx])
		}

		if res.Error != nil && idx < len(tokens) {
			s.logger.Sugar().Warnw("failed to deliver push notification", "error", res.Error, "token", tokens[idx])
		}
	}

	if len(invalidTokens) > 0 {
		if err := s.deviceTokenRepo.DeleteTokens(ctx, invalidTokens); err != nil {
			s.logger.Sugar().Warnw("failed to prune invalid device tokens", "error", err)
		}
	}

	return nil
}

func buildLikeBody(likerName string) string {
	if likerName == "" {
		return "You have a new like waiting for you."
	}

	return fmt.Sprintf("%s likes you. Open Haerd to say hi.", likerName)
}

func buildMatchBody(counterpartName string) string {
	if counterpartName == "" {
		return "You just matched! See who it is."
	}

	return fmt.Sprintf("You and %s just matched. Jump in and start the chat.", counterpartName)
}

func buildMessageTitle(senderName string) string {
	if senderName == "" {
		return "New message"
	}

	return fmt.Sprintf("New message from %s", senderName)
}

func shouldRemoveToken(err error) bool {
	return messaging.IsUnregistered(err) || messaging.IsInvalidArgument(err)
}

var errNoFirebaseCredentials = errors.New("firebase credentials not configured")

func newMessagingClient(ctx context.Context, cfg Config) (*messaging.Client, string, error) {
	var opts []option.ClientOption

	var source string

	if trimmed := strings.TrimSpace(cfg.CredentialsJSON); trimmed != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(trimmed)))
		source = "json"
	} else if cfg.ServiceAccountPath != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.ServiceAccountPath))
		source = "file"
	} else {
		return nil, "", errNoFirebaseCredentials
	}

	fbCfg := &firebase.Config{}
	if cfg.ProjectID != "" {
		fbCfg.ProjectID = cfg.ProjectID
	}

	app, err := firebase.NewApp(ctx, fbCfg, opts...)
	if err != nil {
		return nil, source, fmt.Errorf("create firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, source, fmt.Errorf("create messaging client: %w", err)
	}

	return client, source, nil
}

func durationUntilNextSunday() time.Duration {
	next := nextSundayAtWeeklyReset(time.Now().UTC())
	return time.Until(next)
}

func nextSundayAtWeeklyReset(now time.Time) time.Time {
	target := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		weeklyRefreshHour,
		weeklyRefreshMinute,
		weeklyRefreshSecond,
		0,
		time.UTC,
	)

	daysUntilSunday := (int(time.Sunday) - int(now.Weekday()) + 7) % 7
	target = target.AddDate(0, 0, daysUntilSunday)

	if !target.After(now) {
		target = target.AddDate(0, 0, 7)
	}

	return target
}
