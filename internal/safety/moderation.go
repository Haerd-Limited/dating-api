package safety

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/realtime/dto"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	realtimehub "github.com/Haerd-Limited/dating-api/internal/realtime"
	safetydomain "github.com/Haerd-Limited/dating-api/internal/safety/domain"
	safetymapper "github.com/Haerd-Limited/dating-api/internal/safety/mapper"
	safetystorage "github.com/Haerd-Limited/dating-api/internal/safety/storage"
	userdomain "github.com/Haerd-Limited/dating-api/internal/user/domain"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
)

type moderationSideEffect struct {
	reportedUserID string
	actionType     string
	warningID      string
	warningMessage string
	accountStatus  string
	suspendedUntil *time.Time
	reason         *string
}

func (s *service) GetAccountStatus(ctx context.Context, userID string) (safetydomain.AccountStatusSummary, error) {
	state, err := s.userService.GetAccountGateState(ctx, userID)
	if err != nil {
		return safetydomain.AccountStatusSummary{}, commonlogger.LogError(s.logger, "get account status", err, zap.String("userID", userID))
	}

	now := time.Now().UTC()
	summary := safetydomain.AccountStatusSummary{
		Status:            state.EffectiveStatus(now),
		HasPendingWarning: state.HasPendingWarning,
	}

	if state.SuspendedUntil != nil && summary.Status == userdomain.AccountStatusSuspended {
		until := state.SuspendedUntil.UTC()
		summary.SuspendedUntil = &until
	}

	return summary, nil
}

func (s *service) ListUnacknowledgedWarnings(ctx context.Context, userID string) ([]safetydomain.ModerationWarning, error) {
	warnings, err := s.repo.ListUnacknowledgedWarnings(ctx, userID)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "list unacknowledged warnings", err, zap.String("userID", userID))
	}

	out := make([]safetydomain.ModerationWarning, 0, len(warnings))
	for _, warning := range warnings {
		out = append(out, safetymapper.ModerationWarningEntityToDomain(warning))
	}

	return out, nil
}

func (s *service) AcknowledgeWarning(ctx context.Context, userID, warningID string) error {
	err := s.repo.AcknowledgeWarning(ctx, warningID, userID)
	if err != nil {
		if errors.Is(err, safetystorage.ErrModerationWarningNotFound) {
			return err
		}

		return commonlogger.LogError(s.logger, "acknowledge warning", err, zap.String("userID", userID), zap.String("warningID", warningID))
	}

	return nil
}

func (s *service) applyModerationAction(
	ctx context.Context,
	req safetydomain.ResolveReportRequest,
	reportEntity *entity.UserReport,
	tx *sql.Tx,
) (*moderationSideEffect, error) {
	reportedUserID := reportEntity.ReportedUserID
	reason := moderationReasonFromRequest(req)

	switch req.ActionType {
	case safetydomain.ActionWarnUser:
		message, err := extractWarningMessage(req)
		if err != nil {
			return nil, err
		}

		warning := &entity.UserModerationWarning{
			UserID:   reportedUserID,
			ReportID: null.StringFrom(req.ReportID),
			Message:  message,
		}

		if err := s.repo.InsertModerationWarning(ctx, warning, tx); err != nil {
			return nil, fmt.Errorf("insert moderation warning: %w", err)
		}

		if warning.ID == "" {
			if reloadErr := warning.Reload(ctx, tx); reloadErr != nil {
				return nil, fmt.Errorf("reload moderation warning: %w", reloadErr)
			}
		}

		return &moderationSideEffect{
			reportedUserID: reportedUserID,
			actionType:     req.ActionType,
			warningID:      warning.ID,
			warningMessage: message,
		}, nil

	case safetydomain.ActionSuspendUser:
		until, err := parseSuspendUntil(req.ActionData)
		if err != nil {
			return nil, err
		}

		if err := s.userService.UpdateAccountStatus(ctx, reportedUserID, userdomain.AccountStatusSuspended, until, reason, tx); err != nil {
			return nil, fmt.Errorf("suspend user: %w", err)
		}

		return &moderationSideEffect{
			reportedUserID: reportedUserID,
			actionType:     req.ActionType,
			accountStatus:  userdomain.AccountStatusSuspended,
			suspendedUntil: until,
			reason:         reason,
		}, nil

	case safetydomain.ActionBanUser:
		if err := s.userService.UpdateAccountStatus(ctx, reportedUserID, userdomain.AccountStatusBanned, nil, reason, tx); err != nil {
			return nil, fmt.Errorf("ban user: %w", err)
		}

		return &moderationSideEffect{
			reportedUserID: reportedUserID,
			actionType:     req.ActionType,
			accountStatus:  userdomain.AccountStatusBanned,
			reason:         reason,
		}, nil

	case safetydomain.ActionEscalate, safetydomain.ActionNoAction:
		return nil, nil

	default:
		return nil, nil
	}
}

func (s *service) runPostCommitModerationEffects(ctx context.Context, effect *moderationSideEffect) {
	if effect == nil {
		return
	}

	switch effect.actionType {
	case safetydomain.ActionSuspendUser, safetydomain.ActionBanUser:
		if err := s.authService.RevokeAllUserSessions(ctx, effect.reportedUserID); err != nil {
			s.logger.Warn("failed to revoke sessions after moderation action",
				zap.String("userID", effect.reportedUserID),
				zap.Error(err),
			)
		}

		s.broadcastAccountStatusChanged(effect)
		s.sendAccountStatusPush(ctx, effect)
	case safetydomain.ActionWarnUser:
		s.broadcastModerationWarning(effect)
		s.sendModerationWarningPush(ctx, effect)
	}
}

func (s *service) broadcastAccountStatusChanged(effect *moderationSideEffect) {
	if s.hub == nil {
		return
	}

	payload := map[string]any{
		"status": effect.accountStatus,
	}

	if effect.suspendedUntil != nil {
		payload["until"] = effect.suspendedUntil.UTC().Format(time.RFC3339)
	}

	if effect.reason != nil {
		payload["reason"] = *effect.reason
	}

	s.broadcastUserEvent(effect.reportedUserID, "account.status_changed", payload)
}

func (s *service) broadcastModerationWarning(effect *moderationSideEffect) {
	if s.hub == nil {
		return
	}

	s.broadcastUserEvent(effect.reportedUserID, "moderation.warning", map[string]any{
		"warning_id": effect.warningID,
		"message":    effect.warningMessage,
	})
}

func (s *service) broadcastUserEvent(userID, eventType string, data map[string]any) {
	evt := dto.Event{
		ID:        realtimehub.NewEventID(),
		Type:      eventType,
		ActorID:   "",
		Ts:        time.Now().UTC(),
		ContextID: "",
		Data:      data,
		Version:   1,
	}

	b, err := json.Marshal(evt)
	if err != nil {
		s.logger.Sugar().Warnw("failed to marshal realtime event", "type", eventType, "error", err)
		return
	}

	s.hub.BroadcastToUser(userID, b)
}

func (s *service) sendAccountStatusPush(ctx context.Context, effect *moderationSideEffect) {
	if s.notificationService == nil {
		return
	}

	var err error

	switch effect.accountStatus {
	case userdomain.AccountStatusSuspended:
		if effect.suspendedUntil != nil {
			err = s.notificationService.SendAccountSuspendedNotification(ctx, effect.reportedUserID, *effect.suspendedUntil)
		}
	case userdomain.AccountStatusBanned:
		err = s.notificationService.SendAccountBannedNotification(ctx, effect.reportedUserID)
	}

	if err != nil {
		s.logger.Warn("failed to send account status push",
			zap.String("userID", effect.reportedUserID),
			zap.Error(err),
		)
	}
}

func (s *service) sendModerationWarningPush(ctx context.Context, effect *moderationSideEffect) {
	if s.notificationService == nil {
		return
	}

	if err := s.notificationService.SendModerationWarningNotification(ctx, effect.reportedUserID, effect.warningMessage); err != nil {
		s.logger.Warn("failed to send moderation warning push",
			zap.String("userID", effect.reportedUserID),
			zap.Error(err),
		)
	}
}

func extractWarningMessage(req safetydomain.ResolveReportRequest) (string, error) {
	if req.ActionData != nil {
		if raw, ok := req.ActionData["warning_message"]; ok {
			if message, ok := raw.(string); ok {
				message = strings.TrimSpace(message)
				if message != "" {
					return message, nil
				}
			}
		}
	}

	if req.Notes != nil {
		message := strings.TrimSpace(*req.Notes)
		if message != "" {
			return message, nil
		}
	}

	return "", ErrMissingWarningMessage
}

func parseSuspendUntil(actionData map[string]any) (*time.Time, error) {
	if actionData == nil {
		return nil, ErrInvalidSuspendUntil
	}

	raw, ok := actionData["suspend_until"]
	if !ok {
		return nil, ErrInvalidSuspendUntil
	}

	value, ok := raw.(string)
	if !ok {
		return nil, ErrInvalidSuspendUntil
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return nil, ErrInvalidSuspendUntil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid suspend_until format", ErrInvalidSuspendUntil)
	}

	if !parsed.After(time.Now().UTC()) {
		return nil, ErrInvalidSuspendUntil
	}

	return &parsed, nil
}

func moderationReasonFromRequest(req safetydomain.ResolveReportRequest) *string {
	if req.Notes == nil {
		return nil
	}

	notes := strings.TrimSpace(*req.Notes)
	if notes == "" {
		return nil
	}

	return &notes
}
