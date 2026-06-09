package matchslot

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	realtimeDTO "github.com/Haerd-Limited/dating-api/internal/api/realtime/dto"
	"github.com/Haerd-Limited/dating-api/internal/matchslot/storage"
	"github.com/Haerd-Limited/dating-api/internal/notification"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	"github.com/Haerd-Limited/dating-api/internal/realtime"
)

type Notifier interface {
	NotifySlotFreed(ctx context.Context, userID string)
	NotifySlotFilled(ctx context.Context, userID string)
}

type notifier struct {
	logger              *zap.Logger
	repo                storage.Repository
	hub                 realtime.Broadcaster
	notificationService notification.Service
	profileService      profile.Service
}

func NewNotifier(
	logger *zap.Logger,
	repo storage.Repository,
	hub realtime.Broadcaster,
	notificationService notification.Service,
	profileService profile.Service,
) Notifier {
	return &notifier{
		logger:              logger,
		repo:                repo,
		hub:                 hub,
		notificationService: notificationService,
		profileService:      profileService,
	}
}

func (n *notifier) NotifySlotFreed(ctx context.Context, userID string) {
	n.notify(ctx, userID, false, true)
}

func (n *notifier) NotifySlotFilled(ctx context.Context, userID string) {
	n.notify(ctx, userID, true, false)
}

func (n *notifier) notify(ctx context.Context, userID string, atMatchLimit, sendPush bool) {
	if n == nil || n.repo == nil {
		return
	}

	targets, err := n.repo.GetOutgoingLikeTargetIDs(ctx, userID)
	if err != nil {
		n.logger.Error("matchslot: get outgoing like targets", zap.Error(err), zap.String("userID", userID))
		return
	}

	n.broadcastSlotChange(userID, atMatchLimit, targets)

	if !sendPush {
		return
	}

	favouriters, err := n.repo.GetFavouritersOfUser(ctx, userID)
	if err != nil {
		n.logger.Error("matchslot: get favouriters", zap.Error(err), zap.String("userID", userID))
		return
	}

	if len(favouriters) == 0 {
		return
	}

	displayName := ""

	if n.profileService != nil {
		card, profileErr := n.profileService.GetProfileCard(ctx, userID)
		if profileErr != nil {
			n.logger.Warn("matchslot: get profile card for push", zap.Error(profileErr), zap.String("userID", userID))
		} else {
			displayName = card.DisplayName
		}
	}

	for _, favouriterID := range favouriters {
		if n.notificationService == nil {
			continue
		}

		if pushErr := n.notificationService.SendSlotFreedNotification(ctx, userID, displayName, favouriterID); pushErr != nil {
			n.logger.Warn("matchslot: send slot freed push",
				zap.Error(pushErr),
				zap.String("userID", userID),
				zap.String("favouriterID", favouriterID),
			)
		}
	}
}

func (n *notifier) broadcastSlotChange(userID string, atMatchLimit bool, targets []string) {
	if n.hub == nil || len(targets) == 0 {
		return
	}

	evt := realtimeDTO.Event{
		ID:      realtime.NewEventID(),
		Type:    "match_slot.changed",
		ActorID: userID,
		Ts:      time.Now(),
		Data: map[string]any{
			"user_id":        userID,
			"at_match_limit": atMatchLimit,
		},
		Version: 1,
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		n.logger.Error("matchslot: marshal event", zap.Error(err))
		return
	}

	for _, targetID := range targets {
		n.hub.BroadcastToUser(targetID, payload)
	}
}
