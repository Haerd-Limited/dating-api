package interaction

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/Haerd-Limited/dating-api/internal/matching"
	matchingdomain "github.com/Haerd-Limited/dating-api/internal/matching/domain"
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
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
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
	matchingService     matching.Service
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
	matchingService matching.Service,
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
		matchingService:     matchingService,
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
	ErrUnverifiedUser                          = errors.New("user must be verified before matching")
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

	err = is.validateSwipe(ctx, swipe, matchable)
	if err != nil {
		return "", commonlogger.LogError(is.logger, "validate swipe", err, zap.Any("swipe", swipe))
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

		// Block unverified users from matching
		verified, err := is.profileService.IsVerified(ctx, swipe.UserID)
		if err != nil {
			return "", commonlogger.LogError(is.logger, "check user verification status", err, zap.String("userID", swipe.UserID))
		}

		if !verified {
			return "", ErrUnverifiedUser
		}

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
			messageType := conversationDomain.MessageTypeText

			var textBody *string

			var mediaUrl *string

			var mediaSeconds *float64

			if targetUserSwipe.MessageType.String == constants.MessageTypeVoice {
				messageType = conversationDomain.MessageTypeVoice
				mediaUrl = targetUserSwipe.VoicenoteURL.Ptr()
				// Note: MediaSeconds is not stored in swipes table yet, so we can't retrieve it for existing swipes.
				// A migration is needed to add media_seconds column to the swipes table.
				// For now, this will be nil for existing swipes, which may cause validation errors when creating the message.
				mediaSeconds = nil
			} else {
				textBody = targetUserSwipe.Message.Ptr()
			}

			_, err = is.conversationService.SendMessage(ctx, tx.Raw(), conversationDomain.Message{
				ConversationID: convoID,
				SenderID:       swipe.TargetUserID,
				Type:           messageType,
				TextBody:       textBody,
				MediaUrl:       mediaUrl,
				MediaSeconds:   mediaSeconds,
				ClientMsgID:    targetUserSwipe.IdempotencyKey.String,
			})
			if err != nil {
				return "", commonlogger.LogError(is.logger, "send target user's message to conversation", err, zap.String("userID", swipe.UserID), zap.String("targetUserID", swipe.TargetUserID))
			}
		}

		userRepliedToLiked := swipe.Message != nil || (swipe.MessageType != nil && swipe.VoiceNoteURL != nil)
		if userRepliedToLiked {
			messageType := conversationDomain.MessageTypeText

			var textBody *string

			var mediaUrl *string

			var mediaSeconds *float64

			if swipe.MessageType != nil && *swipe.MessageType == constants.MessageTypeVoice {
				messageType = conversationDomain.MessageTypeVoice
				mediaUrl = swipe.VoiceNoteURL
				mediaSeconds = swipe.MediaSeconds
			} else {
				textBody = swipe.Message
			}

			clientMsgID := ""
			if swipe.IdempotencyKey != nil {
				clientMsgID = *swipe.IdempotencyKey
			}

			_, err = is.conversationService.SendMessage(ctx, tx.Raw(), conversationDomain.Message{
				ConversationID: convoID,
				SenderID:       swipe.UserID,
				Type:           messageType,
				TextBody:       textBody,
				MediaUrl:       mediaUrl,
				MediaSeconds:   mediaSeconds,
				ClientMsgID:    clientMsgID,
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

		p.MatchSummary, likesErr = is.computeMatchSummary(ctx, userID, id)
		if likesErr != nil {
			is.logger.Warn(
				"compute match summary for like",
				zap.Error(likesErr),
				zap.String("userID", userID),
				zap.String("targetUserID", id),
				zap.Int("minOverlap", constants.MatchSummaryMinOverlap),
			)
			p.MatchSummary = nil
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
			like.Prompt.CoverMediaURL = voicePrompt.CoverMediaURL
			like.Prompt.CoverMediaType = voicePrompt.CoverMediaType
			like.Prompt.CoverMediaAspectRatio = voicePrompt.CoverMediaAspectRatio
		}

		if swipe.Message.Valid && swipe.MessageType.Valid {
			like.Message.MessageText, like.Message.MessageType = swipe.Message.Ptr(), swipe.MessageType.Ptr()
		}

		if p.VerifiedStatus == "VERIFIED" {
			likes.Verified = append(likes.Verified, like)
		} else {
			likes.Unverified = append(likes.Unverified, like)
		}
	}

	return likes, nil
}

func (is *service) computeMatchSummary(ctx context.Context, viewerID, targetID string) (*profilecard.MatchSummary, error) {
	matchSummary, err := is.matchingService.ComputeMatch(ctx, viewerID, targetID, constants.MatchSummaryMinOverlap)
	if err != nil {
		return nil, err
	}

	result := &profilecard.MatchSummary{
		MatchPercent: matchSummary.MatchPercent,
		OverlapCount: matchSummary.OverlapCount,
		HiddenReason: matchSummary.HiddenReason,
	}
	result.Badges = mapBadges(matchSummary.Badges)

	return result, nil
}

func mapBadges(badges []matchingdomain.MatchBadge) []profilecard.MatchBadge {
	if len(badges) == 0 {
		return nil
	}

	result := make([]profilecard.MatchBadge, 0, len(badges))
	for _, badge := range badges {
		result = append(result, profilecard.MatchBadge{
			QuestionID:    badge.QuestionID,
			QuestionText:  badge.QuestionText,
			PartnerAnswer: badge.PartnerAnswer,
			Weight:        badge.Weight,
		})
	}

	return result
}

func (is *service) validateSwipe(ctx context.Context, swipe domain.Swipe, isMatchable bool) error {
	// Ensure frontend/client sent a valid action
	if swipe.Action != constants.ActionLike && swipe.Action != constants.ActionSuperlike && swipe.Action != constants.ActionPass {
		return fmt.Errorf("%w : action=%s", ErrInvalidAction, swipe.Action)
	}

	// Check is frontend attempted to send a like with a message but is missing required fields
	if swipe.Action == constants.ActionSuperlike || swipe.Action == constants.ActionLike {
		hasAny := swipe.Message != nil || swipe.MessageType != nil || swipe.IdempotencyKey != nil || swipe.VoiceNoteURL != nil

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

		// Validate voice note requirements if message type is voice
		if swipe.MessageType != nil && *swipe.MessageType == constants.MessageTypeVoice {
			if swipe.VoiceNoteURL == nil || swipe.MediaSeconds == nil {
				return ErrMissingRequiredFieldsForLikeWithMessage
			}
			// Validate media seconds is positive and within limits
			if *swipe.MediaSeconds <= 0 || *swipe.MediaSeconds > float64(constants.MaxVoiceNoteLengthInSeconds) {
				return fmt.Errorf("media_seconds must be between 0 and %d", constants.MaxVoiceNoteLengthInSeconds)
			}
		}
	}

	// block self like
	if swipe.UserID == swipe.TargetUserID {
		return ErrSelfLike
	}

	// trusting frontend to not allow a user to send a simple like to a VWH user. And if the user becomes a VWH user during the time after the user's feed has been fetched, then let them like them

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
