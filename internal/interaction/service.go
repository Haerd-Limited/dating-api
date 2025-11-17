package interaction

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/friendsofgo/errors"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/realtime/dto"
	"github.com/Haerd-Limited/dating-api/internal/conversation"
	conversationDomain "github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	"github.com/Haerd-Limited/dating-api/internal/discover"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/interaction/domain"
	"github.com/Haerd-Limited/dating-api/internal/interaction/mapper"
	"github.com/Haerd-Limited/dating-api/internal/interaction/storage"
	"github.com/Haerd-Limited/dating-api/internal/notification"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	profiledomain "github.com/Haerd-Limited/dating-api/internal/profile/domain"
	"github.com/Haerd-Limited/dating-api/internal/realtime"
	"github.com/Haerd-Limited/dating-api/internal/safety"
	"github.com/Haerd-Limited/dating-api/internal/uow"
	commonanalytics "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/analytics"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
)

type Service interface {
	CreateSwipe(ctx context.Context, swipe domain.Swipe) (string, error)
	GetLikes(ctx context.Context, userID, direction string, offset, limit int) (domain.Likes, error)
}

type service struct {
	logger              *zap.Logger
	profileService      profile.Service
	conversationService conversation.Service
	interactionRepo     storage.InteractionRepository
	discoverService     discover.Service
	safetyService       safety.Service
	uow                 uow.UoW
	hub                 realtime.Broadcaster
	notificationService notification.Service
}

func NewInteractionService(
	logger *zap.Logger,
	profileService profile.Service,
	conversationService conversation.Service,
	interactionRepo storage.InteractionRepository,
	discoverService discover.Service,
	safetyService safety.Service,
	uow uow.UoW,
	hub realtime.Broadcaster,
	notificationService notification.Service,
) Service {
	return &service{
		logger:              logger,
		interactionRepo:     interactionRepo,
		profileService:      profileService,
		conversationService: conversationService,
		discoverService:     discoverService,
		safetyService:       safetyService,
		uow:                 uow,
		hub:                 hub,
		notificationService: notificationService,
	}
}

var (
	ErrSelfLike                                = errors.New("user cannot like themselves")
	ErrPromptIDRequiredToLikeUser              = errors.New("prompt id is required to like a user")
	ErrInvalidDirection                        = errors.New("invalid direction")
	ErrInvalidAction                           = errors.New("invalid action")
	ErrLikedAVhwUser                           = errors.New("user liked a vhw user")
	ErrMissingRequiredFieldsForLikeWithMessage = errors.New("message, message type, prompt id and idempotency key are required for sending a like with a message")
	ErrWeeklySuperlikeLimitReached             = errors.New("weekly superlike limit reached")
	ErrBlockedUser                             = errors.New("cannot interact with a blocked user")
)

const (
	ResultMatched = "MATCHED"
	ResultSent    = "SENT"
	ResultPassed  = "PASSED"
)

const weeklySuperlikeAllowance int64 = 1

func (is *service) CreateSwipe(ctx context.Context, swipe domain.Swipe) (string, error) {
	// this should all be a single transaction
	tx, err := is.uow.Begin(ctx)
	if err != nil {
		return "", commonlogger.LogError(is.logger, "begin tx", err)
	}

	defer func() { _ = tx.Rollback() }()

	matchable, err := is.interactionRepo.CheckIfMatchable(ctx, swipe.UserID, swipe.TargetUserID)
	if err != nil {
		return "", commonlogger.LogError(is.logger, "check if matchable", err, zap.String("userID", swipe.UserID), zap.String("targetUserID", swipe.TargetUserID))
	}

	if is.safetyService != nil {
		blocked, blockErr := is.safetyService.IsBlocked(ctx, swipe.UserID, swipe.TargetUserID)
		if blockErr != nil {
			return "", commonlogger.LogError(is.logger, "check block status", blockErr, zap.String("userID", swipe.UserID), zap.String("targetUserID", swipe.TargetUserID))
		}

		if blocked {
			return "", ErrBlockedUser
		}
	}

	// todo(high-priority): allow voice note swipes
	err = is.validateSwipe(ctx, swipe, matchable)
	if err != nil {
		return "", commonlogger.LogError(is.logger, "validate swipe", err, zap.String("userID", swipe.UserID))
	}

	switch swipe.Action {
	case constants.ActionLike, constants.ActionSuperlike:
		if swipe.Action == constants.ActionSuperlike {
			superlikeWindowStart := startOfWeek(time.Now(), time.Sunday, time.UTC)

			superlikesUsed, err := is.interactionRepo.CountSuperlikesSince(ctx, swipe.UserID, superlikeWindowStart, tx.Raw())
			if err != nil {
				return "", commonlogger.LogError(is.logger, "get weekly superlike usage", err, zap.String("userID", swipe.UserID))
			}

			if superlikesUsed >= weeklySuperlikeAllowance {
				return "", ErrWeeklySuperlikeLimitReached
			}
		}

		if !matchable {
			if swipe.Message == nil {
				systemMsg := messages.LikedYourPromptMsg
				swipe.Message = &systemMsg
				systemMessageType := string(conversationDomain.MessageTypeSystem)
				swipe.MessageType = &systemMessageType
			}

			err = is.interactionRepo.InsertSwipe(ctx, mapper.SwipeToEntity(swipe), tx.Raw())
			if err != nil {
				return "", commonlogger.LogError(is.logger, "insert swipe", err, zap.String("userID", swipe.UserID))
			}

			// analytics: swipe created (non-match path)
			props := map[string]any{
				"action":    swipe.Action,
				"target_id": swipe.TargetUserID,
				"prompt_id": swipe.PromptID,
				"message":   swipe.Message,
				"message_type": func() any {
					if swipe.MessageType == nil {
						return nil
					}
					return *swipe.MessageType
				}(),
			}
			commonanalytics.Track(ctx, "interaction.swipe_created", &swipe.UserID, nil, props)

			err = tx.Commit()
			if err != nil {
				return "", commonlogger.LogError(is.logger, "commit tx", err)
			}

			evt := dto.Event{
				ID:        realtime.NewEventID(),
				Type:      "like.created",
				ActorID:   swipe.UserID,
				Ts:        time.Now(),
				ContextID: "",
				Data:      nil,
				Version:   1,
			}

			b, mErr := json.Marshal(evt)
			if mErr != nil {
				is.logger.Error("marshal event", zap.Error(mErr))
				return ResultSent, nil
			}

			is.hub.BroadcastToUser(swipe.TargetUserID, b)

			is.sendLikeNotification(ctx, swipe.UserID, swipe.TargetUserID)

			return ResultSent, nil
		}

		// todo(high-priority): block unverified users from matching

		err = is.interactionRepo.InsertSwipe(ctx, mapper.SwipeToEntity(swipe), tx.Raw())
		if err != nil {
			return "", fmt.Errorf("insert swipe userID=%s : %w", swipe.UserID, err)
		}

		// analytics: swipe created (potential match path)
		props := map[string]any{
			"action":    swipe.Action,
			"target_id": swipe.TargetUserID,
			"prompt_id": swipe.PromptID,
			"message":   swipe.Message,
			"message_type": func() any {
				if swipe.MessageType == nil {
					return nil
				}
				return *swipe.MessageType
			}(),
		}
		commonanalytics.Track(ctx, "interaction.swipe_created", &swipe.UserID, nil, props)

		// Create a match (normalize order to keep uniqueness deterministic)
		a, b := swipe.UserID, swipe.TargetUserID
		if b < a {
			a, b = b, a
		}

		err = is.interactionRepo.CreateMatch(ctx, entity.Match{UserA: a, UserB: b}, tx.Raw())
		if err != nil {
			return "", commonlogger.LogError(is.logger, "create match", err, zap.String("userID", swipe.UserID), zap.String("targetUserID", swipe.TargetUserID))
		}

		var convoID string

		convoID, err = is.conversationService.CreateConversationViaTx(ctx, swipe.UserID, swipe.TargetUserID, tx.Raw())
		if err != nil {
			return "", commonlogger.LogError(is.logger, "create conversation", err, zap.String("userID", swipe.UserID), zap.String("targetUserID", swipe.TargetUserID))
		}

		var targetUserSwipe *entity.Swipe

		targetUserSwipe, err = is.interactionRepo.GetSwipeByActorIDAndTargetID(ctx, swipe.TargetUserID, swipe.UserID)
		if err != nil {
			return "", commonlogger.LogError(is.logger, "get swipe by actorID and targetID", err, zap.String("actorID", swipe.TargetUserID), zap.String("targetUserID", swipe.UserID))
		}

		targetUserSentMeALikeWithAMessage := targetUserSwipe.Message.Valid && targetUserSwipe.MessageType.Valid && targetUserSwipe.IdempotencyKey.Valid
		if targetUserSentMeALikeWithAMessage {
			// if the user i'm about to match, sent me a like with a message, we want to include that message as the first message in our conversation.
			_, err = is.conversationService.SendMessage(ctx, tx.Raw(), conversationDomain.Message{
				ConversationID: convoID,
				SenderID:       swipe.TargetUserID,
				Type:           conversationDomain.MessageTypeText, // todo(high-priority): update later to be dynamic and check if they sent a voice note message as a like.
				TextBody:       targetUserSwipe.Message.Ptr(),
				MediaUrl:       nil,
				MediaSeconds:   nil,
				ClientMsgID:    targetUserSwipe.IdempotencyKey.String,
			})
			if err != nil {
				return "", commonlogger.LogError(is.logger, "send target user's message to conversation", err, zap.String("userID", swipe.UserID), zap.String("targetUserID", swipe.TargetUserID))
			}
		}

		userRepliedToLiked := swipe.Message != nil
		if userRepliedToLiked {
			_, err = is.conversationService.SendMessage(ctx, tx.Raw(), conversationDomain.Message{
				ConversationID: convoID,
				SenderID:       swipe.UserID,
				Type:           conversationDomain.MessageTypeText, // todo(high-priority): update later to be dynamic and check if they sent a voice note message as a like.
				TextBody:       swipe.Message,
				MediaUrl:       nil,
				MediaSeconds:   nil,
				ClientMsgID:    targetUserSwipe.IdempotencyKey.String,
			})
			if err != nil {
				return "", commonlogger.LogError(is.logger, "user reply to like", err, zap.String("userID", swipe.UserID), zap.String("targetUserID", swipe.TargetUserID))
			}
		}

		err = tx.Commit()
		if err != nil {
			return "", fmt.Errorf("commit tx: %w", err)
		}

		evt := dto.Event{
			ID:        realtime.NewEventID(),
			Type:      "match.created",
			ActorID:   swipe.UserID,
			Ts:        time.Now(),
			ContextID: convoID,
			Data:      nil,
			Version:   1,
		}

		byts, mErr := json.Marshal(evt)
		if mErr != nil {
			is.logger.Error("marshal event", zap.Error(mErr))
			return ResultMatched, nil
		}

		is.hub.BroadcastToUser(swipe.TargetUserID, byts)
		is.hub.BroadcastToUser(swipe.UserID, byts)

		is.sendMatchNotifications(ctx, swipe.UserID, swipe.TargetUserID, convoID)

		// analytics: match created
		commonanalytics.Track(ctx, "interaction.match_created", &swipe.UserID, nil, map[string]any{
			"match_id": convoID,
		})

		return ResultMatched, nil

	case constants.ActionPass:
		err = is.interactionRepo.InsertSwipe(ctx, mapper.SwipeToEntity(swipe), tx.Raw())
		if err != nil {
			return "", commonlogger.LogError(is.logger, "insert swipe", err, zap.String("userID", swipe.UserID))
		}

		// analytics: pass swipe
		commonanalytics.Track(ctx, "interaction.swipe_created", &swipe.UserID, nil, map[string]any{
			"action": "pass",
		})

		err = tx.Commit()
		if err != nil {
			return "", commonlogger.LogError(is.logger, "commit tx", err)
		}

		return ResultPassed, nil
	}

	return "", fmt.Errorf("%w: %s", ErrInvalidAction, swipe.Action)
}

func (is *service) GetLikes(ctx context.Context, userID, direction string, offset, limit int) (domain.Likes, error) {
	var likesUserIDs []string

	var err error

	switch direction {
	case "incoming":
		likesUserIDs, err = is.interactionRepo.GetIncomingLikes(ctx, userID, limit, offset)
	default:
		return domain.Likes{}, ErrInvalidDirection
	}

	if err != nil {
		return domain.Likes{}, err
	}

	var likes domain.Likes

	for _, id := range likesUserIDs {
		if is.safetyService != nil {
			isBlocked, blockErr := is.safetyService.IsBlocked(ctx, userID, id)
			if blockErr != nil {
				return domain.Likes{}, commonlogger.LogError(is.logger, "check block status", blockErr, zap.String("userID", userID), zap.String("targetUserID", id))
			}

			if isBlocked {
				continue
			}
		}

		alreadyMatched, likesErr := is.interactionRepo.AlreadyMatched(ctx, userID, id)
		if likesErr != nil {
			return domain.Likes{}, commonlogger.LogError(is.logger, "check if already matched", likesErr, zap.String("userID", userID), zap.String("targetUserID", id))
		}

		if alreadyMatched {
			continue
		}

		p, likesErr := is.profileService.GetProfileCard(ctx, id)
		if likesErr != nil {
			return domain.Likes{}, commonlogger.LogError(is.logger, "get profile card", likesErr, zap.String("userID", userID), zap.String("profileUserID", id))
		}

		swipe, likesErr := is.interactionRepo.GetSwipeByActorIDAndTargetID(ctx, id, userID)
		if likesErr != nil {
			return domain.Likes{}, commonlogger.LogError(is.logger, "get swipe by actorID and targetID", likesErr, zap.String("userID", userID), zap.String("targetUserID", id))
		}

		like := domain.Like{
			Profile: p,
			Message: &domain.Message{},
			Prompt:  &domain.Prompt{},
		}

		var voicePrompt profiledomain.VoicePrompt
		if swipe.PromptID.Valid {
			voicePrompt, likesErr = is.profileService.GetVoicePromptByID(ctx, swipe.PromptID.Int64)
			if likesErr != nil {
				return domain.Likes{}, commonlogger.LogError(is.logger, "get voice prompt by ID", likesErr, zap.String("userID", userID), zap.String("targetUserID", id))
			}

			like.Prompt.PromptID = voicePrompt.PromptID
			like.Prompt.Prompt = voicePrompt.Prompt
			like.Prompt.VoiceNoteURL = voicePrompt.VoiceNoteURL
			like.Prompt.CoverPhotoUrl = voicePrompt.CoverPhotoUrl
		}

		if swipe.Message.Valid && swipe.MessageType.Valid {
			like.Message.MessageText, like.Message.MessageType = swipe.Message.Ptr(), swipe.MessageType.Ptr()
		}

		if p.Verified {
			likes.Verified = append(likes.Verified, like)
		} else {
			likes.Unverified = append(likes.Unverified, like)
		}
	}

	return likes, nil
}

func (is *service) validateSwipe(ctx context.Context, swipe domain.Swipe, isMatchable bool) error {
	// Ensure frontend/client sent a valid action
	if swipe.Action != constants.ActionLike && swipe.Action != constants.ActionSuperlike && swipe.Action != constants.ActionPass {
		return fmt.Errorf("%w : action=%s", ErrInvalidAction, swipe.Action)
	}

	// Check is frontend attempted to send a like with a message but is missing required fields
	if swipe.Action == constants.ActionSuperlike || swipe.Action == constants.ActionLike {
		hasAny := swipe.Message != nil || swipe.MessageType != nil || swipe.IdempotencyKey != nil

		var atleastOneIsMissing bool
		if isMatchable {
			atleastOneIsMissing = swipe.Message == nil || swipe.MessageType == nil || swipe.IdempotencyKey == nil
		} else {
			// Check if promptID is nil or 0 (0 is not a valid prompt_id)
			hasValidPromptID := swipe.PromptID != nil && *swipe.PromptID != 0
			atleastOneIsMissing = swipe.Message == nil || swipe.MessageType == nil || !hasValidPromptID || swipe.IdempotencyKey == nil
		}

		unableToSendLikeWithMessage := hasAny && atleastOneIsMissing

		if unableToSendLikeWithMessage {
			return ErrMissingRequiredFieldsForLikeWithMessage
		}
	}

	// block self like
	if swipe.UserID == swipe.TargetUserID {
		return ErrSelfLike
	}

	// Ensure that a user can only superlike or pass a vwh user
	vwhIDs, err := is.discoverService.GetVoiceWorthHearingIDs(ctx, swipe.UserID)
	if err != nil {
		return commonlogger.LogError(is.logger, "get voice worth hearing ids", err, zap.String("userID", swipe.UserID))
	}

	userLikedAVwhUser := len(vwhIDs) != 0 && swipe.Action == constants.ActionLike && slices.Contains(vwhIDs, swipe.TargetUserID)
	if userLikedAVwhUser {
		return ErrLikedAVhwUser
	}

	// If User is sending a like/superlike to someone who hasn't liked them back, then you must provide a promptID.
	alreadyInteracted, err := is.discoverService.AlreadyInteracted(ctx, swipe.UserID, swipe.TargetUserID)
	if err != nil {
		return commonlogger.LogError(is.logger, "already interacted", err, zap.String("userID", swipe.UserID), zap.String("targetUserID", swipe.TargetUserID))
	}

	// Check if promptID is nil or 0 (0 is not a valid prompt_id)
	hasValidPromptID := swipe.PromptID != nil && *swipe.PromptID != 0

	sendingFirstLikeWithoutPromptID := (!alreadyInteracted && swipe.Action == constants.ActionLike && !hasValidPromptID) || (!alreadyInteracted && swipe.Action == constants.ActionSuperlike && !hasValidPromptID)
	if sendingFirstLikeWithoutPromptID {
		return ErrPromptIDRequiredToLikeUser
	}

	return nil
}

func startOfWeek(t time.Time, weekStart time.Weekday, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}

	t = t.In(loc)

	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	offset := (int(midnight.Weekday()) - int(weekStart) + 7) % 7

	return midnight.AddDate(0, 0, -offset)
}

func (is *service) sendLikeNotification(ctx context.Context, actorID, recipientID string) {
	if is.notificationService == nil {
		return
	}

	profile, err := is.profileService.GetProfileCard(ctx, actorID)
	if err != nil {
		is.logger.Sugar().Warnw("send like notification: get profile", "error", err, "userID", actorID)
		return
	}

	if err := is.notificationService.SendLikeNotification(ctx, actorID, profile.DisplayName, recipientID); err != nil {
		is.logger.Sugar().Warnw("failed to send like notification", "error", err, "actorID", actorID, "recipientID", recipientID)
	}
}

func (is *service) sendMatchNotifications(ctx context.Context, userA, userB, conversationID string) {
	if is.notificationService == nil {
		return
	}

	userAProfile, err := is.profileService.GetProfileCard(ctx, userA)
	if err != nil {
		is.logger.Sugar().Warnw("send match notification: get profile", "error", err, "userID", userA)
		return
	}

	userBProfile, err := is.profileService.GetProfileCard(ctx, userB)
	if err != nil {
		is.logger.Sugar().Warnw("send match notification: get profile", "error", err, "userID", userB)
		return
	}

	if err := is.notificationService.SendMatchNotification(ctx, userBProfile.DisplayName, userA, conversationID); err != nil {
		is.logger.Sugar().Warnw("failed to send match notification", "error", err, "recipientID", userA, "counterpartID", userB)
	}

	if err := is.notificationService.SendMatchNotification(ctx, userAProfile.DisplayName, userB, conversationID); err != nil {
		is.logger.Sugar().Warnw("failed to send match notification", "error", err, "recipientID", userB, "counterpartID", userA)
	}
}
