package interaction

import (
	"context"
	"fmt"
	storage2 "github.com/Haerd-Limited/dating-api/internal/discover/storage"
	"slices"

	"github.com/friendsofgo/errors"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/conversation"
	conversationDomain "github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/interaction/domain"
	"github.com/Haerd-Limited/dating-api/internal/interaction/mapper"
	"github.com/Haerd-Limited/dating-api/internal/interaction/storage"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	profiledomain "github.com/Haerd-Limited/dating-api/internal/profile/domain"
	"github.com/Haerd-Limited/dating-api/internal/uow"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
)

type Service interface {
	CreateSwipe(ctx context.Context, swipe domain.Swipe) (string, error)
	GetLikes(ctx context.Context, userID, direction string, offset, limit int) ([]domain.Like, error)
}

type service struct {
	logger              *zap.Logger
	profileService      profile.Service
	conversationService conversation.Service
	interactionRepo     storage.InteractionRepository
	discoverRepo        storage2.DiscoverRepository
	uow                 uow.UoW
}

func NewInteractionService(
	logger *zap.Logger,
	profileService profile.Service,
	conversationService conversation.Service,
	interactionRepo storage.InteractionRepository,
	discoverRepo storage2.DiscoverRepository,
	uow uow.UoW,
) Service {
	return &service{
		logger:              logger,
		interactionRepo:     interactionRepo,
		profileService:      profileService,
		conversationService: conversationService,
		uow:                 uow,
		discoverRepo:        discoverRepo,
	}
}

var (
	ErrInvalidDirection                        = errors.New("invalid direction")
	ErrInvalidAction                           = errors.New("invalid action")
	ErrLikedAVhwUser                           = errors.New("user did not super like a voice worth hearing user")
	ErrMissingRequiredFieldsForLikeWithMessage = errors.New("message, message type, prompt id and idempotency key are required for like with message")
)

const (
	ResultMatched = "MATCHED"
	ResultSent    = "SENT"
	ResultPassed  = "PASSED"
)

// todo: implement only one superlike a week unless they buy more.
func (is *service) CreateSwipe(ctx context.Context, swipe domain.Swipe) (string, error) {
	// this should all be a single transaction
	tx, err := is.uow.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	err = is.validateSwipe(ctx, swipe)
	if err != nil {
		return "", fmt.Errorf("validate swipe userID=%s : %w", swipe.UserID, err)
	}

	switch swipe.Action {
	case constants.ActionLike, constants.ActionSuperlike:
		err = is.interactionRepo.InsertSwipe(ctx, mapper.SwipeToEntity(swipe), tx.Raw())
		if err != nil {
			return "", fmt.Errorf("insert swipe userID=%s : %w", swipe.UserID, err)
		}

		var matchErr error

		matchable, matchErr := is.interactionRepo.CheckIfMatchable(ctx, swipe.UserID, swipe.TargetUserID)
		if matchErr != nil {
			return "", fmt.Errorf("check if matchable userID=%s targetUserID=%s: %w", swipe.UserID, swipe.TargetUserID, matchErr)
		}

		if !matchable {
			err = tx.Commit()
			if err != nil {
				return "", fmt.Errorf("commit tx: %w", err)
			}

			return ResultSent, nil
		}
		// Create a match (normalize order to keep uniqueness deterministic)
		a, b := swipe.UserID, swipe.TargetUserID
		if b < a {
			a, b = b, a
		}

		err = is.interactionRepo.CreateMatch(ctx, entity.Match{
			UserA: a,
			UserB: b,
		},
			tx.Raw(),
		)
		if err != nil {
			return "", fmt.Errorf("create match userID=%s targetUserID=%s: %w", swipe.UserID, swipe.TargetUserID, err)
		}

		var convoID string

		convoID, err = is.conversationService.CreateConversationViaTx(ctx, swipe.UserID, swipe.TargetUserID, tx.Raw())
		if err != nil {
			return "", fmt.Errorf("create conversation userID=%s targetUserID=%s: %w", swipe.UserID, swipe.TargetUserID, err)
		}

		var targetUserSwipe *entity.Swipe

		targetUserSwipe, err = is.interactionRepo.GetSwipeByActorIDAndTargetID(ctx, swipe.TargetUserID, swipe.UserID)
		if err != nil {
			return "", fmt.Errorf("get swipe by actorID and targetID actorID=%s targetUserID=%s: %w", swipe.TargetUserID, swipe.UserID, err)
		}

		targetUserSentMeALikeWithAMessage := targetUserSwipe.Message.Valid && targetUserSwipe.MessageType.Valid && targetUserSwipe.IdempotencyKey.Valid
		if targetUserSentMeALikeWithAMessage {
			// if the user i'm about to match, sent me a like with a message, we want to include that message as the first message in our conversation.
			_, err = is.conversationService.SendMessageViaTx(ctx, tx.Raw(), conversationDomain.Message{
				ConversationID: convoID,
				SenderID:       swipe.TargetUserID,
				Type:           conversationDomain.MessageTypeText, // todo: update later to be dynamic and check if they sent a voice note message as a like.
				TextBody:       targetUserSwipe.Message.Ptr(),
				MediaKey:       nil,
				MediaSeconds:   nil,
				ClientMsgID:    targetUserSwipe.IdempotencyKey.String,
			})
			if err != nil {
				return "", fmt.Errorf("send message userID=%s targetUserID=%s: %w", swipe.UserID, swipe.TargetUserID, err)
			}
		}

		err = tx.Commit()
		if err != nil {
			return "", fmt.Errorf("commit tx: %w", err)
		}

		return ResultMatched, nil

	case constants.ActionPass:
		err = is.interactionRepo.InsertSwipe(ctx, mapper.SwipeToEntity(swipe), tx.Raw())
		if err != nil {
			return "", fmt.Errorf("insert swipe userID=%s : %w", swipe.UserID, err)
		}

		err = tx.Commit()
		if err != nil {
			return "", fmt.Errorf("commit tx: %w", err)
		}

		return ResultPassed, nil
	}

	return "", fmt.Errorf("invalid action: %s", swipe.Action)
}

func (is *service) GetLikes(ctx context.Context, userID, direction string, offset, limit int) ([]domain.Like, error) {
	var likesUserIDs []string

	var err error

	switch direction {
	case "incoming":
		likesUserIDs, err = is.interactionRepo.GetIncomingLikes(ctx, userID, limit, offset)
	default:
		return nil, ErrInvalidDirection
	}

	if err != nil {
		return nil, err
	}

	var likes []domain.Like

	for _, id := range likesUserIDs {
		alreadyMatched, likesErr := is.interactionRepo.AlreadyMatched(ctx, userID, id)
		if likesErr != nil {
			return nil, fmt.Errorf("check if already matched userID=%s targetUserID=%s: %w", userID, id, likesErr)
		}

		if alreadyMatched {
			continue
		}

		p, likesErr := is.profileService.GetProfileCard(ctx, id)
		if likesErr != nil {
			return nil, fmt.Errorf("get profile card userID=%s profileUserID=%s: %w", userID, id, likesErr)
		}

		swipe, likesErr := is.interactionRepo.GetSwipeByActorIDAndTargetID(ctx, id, userID)
		if likesErr != nil {
			return nil, fmt.Errorf("get swipe by actorID and targetID userID=%s targetUserID=%s: %w", userID, id, likesErr)
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
				return nil, fmt.Errorf("get voice prompt by ID userID=%s targetUserID=%s: %w", userID, id, likesErr)
			}

			like.Prompt.PromptID = voicePrompt.PromptID
			like.Prompt.Prompt = voicePrompt.Prompt
			like.Prompt.VoiceNoteURL = voicePrompt.VoiceNoteURL
			like.Prompt.CoverPhotoUrl = voicePrompt.CoverPhotoUrl
		}

		if swipe.Message.Valid && swipe.MessageType.Valid {
			like.Message.MessageText, like.Message.MessageType = swipe.Message.Ptr(), swipe.MessageType.Ptr()
		}

		likes = append(likes, like)
	}

	return likes, nil
}

func (is *service) validateSwipe(ctx context.Context, swipe domain.Swipe) error {
	//Ensure frontend/client sent a valid action
	if swipe.Action != constants.ActionLike && swipe.Action != constants.ActionSuperlike && swipe.Action != constants.ActionPass {
		return fmt.Errorf("%w : action=%s", ErrInvalidAction, swipe.Action)
	}

	//Check is frontend attempted to send a like with a message but is missing required fields
	if swipe.Action == constants.ActionSuperlike || swipe.Action == constants.ActionLike {
		MissingIdempotencyKey := swipe.Message != nil && swipe.MessageType != nil && swipe.PromptID != nil && swipe.IdempotencyKey == nil
		MissingPromptID := swipe.Message != nil && swipe.MessageType != nil && swipe.PromptID == nil && swipe.IdempotencyKey != nil
		MissingMessageType := swipe.Message != nil && swipe.MessageType == nil && swipe.PromptID != nil && swipe.IdempotencyKey != nil
		MissingMessage := swipe.Message == nil && swipe.MessageType != nil && swipe.PromptID != nil && swipe.IdempotencyKey == nil
		unableToSendLikeWithMessage := MissingIdempotencyKey || MissingPromptID || MissingMessageType || MissingMessage
		if unableToSendLikeWithMessage {
			return ErrMissingRequiredFieldsForLikeWithMessage
		}
	}

	//Ensure that a user can only superlike a vwh user
	vwhIDs, err := is.discoverRepo.GetVoiceWorthHearingIDs(ctx, swipe.UserID)
	if err != nil {
		return fmt.Errorf("get voice worth hearing ids userID=%s: %w", swipe.UserID, err)
	}
	userDidNotSuperLikeAVwhUser := len(vwhIDs) != 0 && swipe.Action == constants.ActionLike && slices.Contains(vwhIDs, swipe.TargetUserID)
	if userDidNotSuperLikeAVwhUser {
		return ErrLikedAVhwUser
	}

	return nil
}
