package notification

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/notification/domain"
	"github.com/Haerd-Limited/dating-api/internal/notification/storage"
)

const (
	weeklyRefreshHour   = 0
	weeklyRefreshMinute = 5
	weeklyRefreshSecond = 0
	weeklyBatchSize     = 200
	expoAPIEndpoint     = "https://exp.host/--/api/v2/push/send"
)

type Service interface {
	RegisterDeviceToken(ctx context.Context, userID, token string) error
	RemoveDeviceToken(ctx context.Context, userID, token string) error
	SendLikeNotification(ctx context.Context, likerID, likerName, recipientUserID string) error
	SendMatchNotification(ctx context.Context, counterpartName, recipientUserID, conversationID string) error
	SendNewMessageNotification(ctx context.Context, senderID, senderName, conversationID, recipientUserID, preview string) error
	SendVerificationApprovedNotification(ctx context.Context, recipientUserID string) error
	SendVerificationRejectedNotification(ctx context.Context, recipientUserID, rejectionReason string) error
	SendRevealRequestNotification(ctx context.Context, initiatorID, initiatorName, recipientUserID, conversationID string) error
	SendRevealAcceptedNotification(ctx context.Context, acceptorID, acceptorName, recipientUserID, conversationID string) error
	SendWeeklyRefreshNotifications(ctx context.Context) error
	StartWeeklyRefreshScheduler(ctx context.Context)
}

type service struct {
	logger            *zap.Logger
	deviceTokenRepo   storage.DeviceTokenRepository
	httpClient        *http.Client
	expoAccessToken   string
	weeklyJobStarted  atomic.Bool
	disableMessaging  bool
	credentialsSource string
}

type Config struct {
	ExpoAccessToken string
}

type expoRequest struct {
	To    string            `json:"to"`
	Sound string            `json:"sound"`
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data"`
}

type expoResponse struct {
	Data struct {
		Status string `json:"status"`
		ID     string `json:"id"`
	} `json:"data"`
}

type expoError struct {
	code    string
	message string
}

func (e *expoError) Error() string {
	return fmt.Sprintf("expo error [%s]: %s", e.code, e.message)
}

func NewService(ctx context.Context, logger *zap.Logger, repo storage.DeviceTokenRepository, cfg Config) (Service, error) {
	token := strings.TrimSpace(cfg.ExpoAccessToken)

	if token == "" {
		logger.Sugar().Warn("expo access token not provided; push notifications disabled but device token APIs will continue to work")
		return &service{
			logger:            logger,
			deviceTokenRepo:   repo,
			httpClient:        nil,
			disableMessaging:  true,
			credentialsSource: "",
		}, nil
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	logger.Sugar().Info("expo push notification client initialized")

	return &service{
		logger:            logger,
		deviceTokenRepo:   repo,
		httpClient:        httpClient,
		expoAccessToken:   token,
		credentialsSource: "env",
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
			"type":          "like.new",
			"liker_id":      likerID,
			"liker_name":    likerName,
			"recipient_id":  recipientUserID,
			"timestamp_utc": time.Now().UTC().Format(time.RFC3339),
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
		},
	}

	return s.sendToUsers(ctx, []string{recipientUserID}, msg)
}

func (s *service) SendVerificationApprovedNotification(ctx context.Context, recipientUserID string) error {
	msg := domain.Message{
		Title: "Verification approved",
		Body:  "Your video verification has been approved! Your profile is now verified.",
		Data: map[string]string{
			"type":          "verification.approved",
			"recipient_id":  recipientUserID,
			"timestamp_utc": time.Now().UTC().Format(time.RFC3339),
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
		},
	}

	return s.sendToUsers(ctx, []string{recipientUserID}, msg)
}

func (s *service) SendRevealRequestNotification(ctx context.Context, initiatorID, initiatorName, recipientUserID, conversationID string) error {
	body := buildRevealRequestBody(initiatorName)
	msg := domain.Message{
		Title: "Reveal request",
		Body:  body,
		Data: map[string]string{
			"type":            "reveal.request",
			"conversation_id": conversationID,
			"initiator_id":    initiatorID,
			"initiator_name":  initiatorName,
			"recipient_id":    recipientUserID,
			"timestamp_utc":   time.Now().UTC().Format(time.RFC3339),
		},
	}

	return s.sendToUsers(ctx, []string{recipientUserID}, msg)
}

func (s *service) SendRevealAcceptedNotification(ctx context.Context, acceptorID, acceptorName, recipientUserID, conversationID string) error {
	body := buildRevealAcceptedBody(acceptorName)
	msg := domain.Message{
		Title: "Reveal accepted",
		Body:  body,
		Data: map[string]string{
			"type":            "reveal.accepted",
			"conversation_id": conversationID,
			"acceptor_id":     acceptorID,
			"acceptor_name":   acceptorName,
			"recipient_id":    recipientUserID,
			"timestamp_utc":   time.Now().UTC().Format(time.RFC3339),
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
			"type":          "weekly.refresh",
			"timestamp_utc": time.Now().UTC().Format(time.RFC3339),
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
	if s.httpClient == nil {
		s.logger.Debug("push messaging disabled; skipping send")
		return nil
	}

	var invalidTokens []string
	var sendErr error

	for _, token := range tokens {
		if err := s.sendToExpo(ctx, token, msg); err != nil {
			if isInvalidTokenError(err) {
				invalidTokens = append(invalidTokens, token)
			}
			sendErr = multierr.Append(sendErr, fmt.Errorf("send to token %s: %w", hashToken(token), err))
			s.logger.Sugar().Warnw("failed to deliver push notification", "error", err, "token_hash", hashToken(token))
		}
	}

	if len(invalidTokens) > 0 {
		if err := s.deviceTokenRepo.DeleteTokens(ctx, invalidTokens); err != nil {
			s.logger.Sugar().Warnw("failed to prune invalid device tokens", "error", err)
		}
	}

	return sendErr
}

func (s *service) sendToExpo(ctx context.Context, token string, msg domain.Message) error {
	reqBody := expoRequest{
		To:    token,
		Sound: "default",
		Title: msg.Title,
		Body:  msg.Body,
		Data:  msg.Data,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", expoAPIEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.expoAccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip, deflate")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expo api error (status %d): %s", resp.StatusCode, string(body))
	}

	var expoResp expoResponse
	if err := json.NewDecoder(resp.Body).Decode(&expoResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if expoResp.Data.Status == "error" {
		return &expoError{code: "unknown", message: "expo returned error status"}
	}

	return nil
}

func isInvalidTokenError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "devicenotregistered") ||
		strings.Contains(errStr, "invalidexpopushtoken") ||
		strings.Contains(errStr, "status 400")
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:8])
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

func buildRevealRequestBody(initiatorName string) string {
	if initiatorName == "" {
		return "You have a reveal request waiting for you."
	}

	return fmt.Sprintf("%s wants to reveal photos. Open Haerd to respond.", initiatorName)
}

func buildRevealAcceptedBody(acceptorName string) string {
	if acceptorName == "" {
		return "Your reveal request was accepted! See who it is."
	}

	return fmt.Sprintf("%s accepted your reveal request. See their photos now.", acceptorName)
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
