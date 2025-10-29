package conversation

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/conversation/dto"
	"github.com/Haerd-Limited/dating-api/internal/api/conversation/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/conversation"
	"github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	convostorage "github.com/Haerd-Limited/dating-api/internal/conversation/storage"
	"github.com/Haerd-Limited/dating-api/internal/interaction"
	interactionstorage "github.com/Haerd-Limited/dating-api/internal/interaction/storage"
	"github.com/Haerd-Limited/dating-api/internal/user/storage"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	commonMessages "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	GetConversations() http.HandlerFunc
	SendMessage() http.HandlerFunc
	GetConversationMessages() http.HandlerFunc
	GetConversationScore() http.HandlerFunc
	InitiateReveal() http.HandlerFunc
	ConfirmReveal() http.HandlerFunc
	MakeRevealDecision() http.HandlerFunc
	GetMatchPhotos() http.HandlerFunc
}

type handler struct {
	logger              *zap.Logger
	conversationService conversation.Service
}

func NewConversationHandler(
	logger *zap.Logger,
	conversationService conversation.Service,
) Handler {
	return &handler{
		logger:              logger,
		conversationService: conversationService,
	}
}

func (h *handler) InitiateReveal() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		convoID := chi.URLParam(r, "id")
		if convoID == "" {
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(
					"Conversation ID is required as a URL parameter",
				))

			return
		}

		err := h.conversationService.InitiateReveal(ctx, userID, convoID)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "InitiateReveal", err)
			return
		}

		render.Json(w, http.StatusOK, dto.InitiateRevealResponse{
			Message: "Reveal request initiated successfully",
		})
	}
}

func (h *handler) ConfirmReveal() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		convoID := chi.URLParam(r, "id")
		if convoID == "" {
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(
					"Conversation ID is required as a URL parameter",
				))

			return
		}

		err := h.conversationService.ConfirmReveal(ctx, userID, convoID)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "ConfirmReveal", err)
			return
		}

		// Get photos after reveal
		photos, err := h.conversationService.GetMatchPhotos(ctx, convoID, userID)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetMatchPhotos", err)
			return
		}

		render.Json(w, http.StatusOK, dto.ConfirmRevealResponse{
			Photos: mapper.MapPhotosToDTO(photos),
		})
	}
}

func (h *handler) MakeRevealDecision() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		convoID := chi.URLParam(r, "id")
		if convoID == "" {
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(
					"Conversation ID is required as a URL parameter",
				))

			return
		}

		var req dto.MakeRevealDecisionRequest

		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate make reveal decision request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse("decision is required and must be one of: continue, date, unmatch"),
			)

			return
		}

		err = h.conversationService.MakeRevealDecision(ctx, userID, convoID, req.Decision)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "MakeRevealDecision", err)
			return
		}

		render.Json(w, http.StatusOK, dto.MakeRevealDecisionResponse{
			Message: "Decision saved successfully",
		})
	}
}

func (h *handler) GetMatchPhotos() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		convoID := chi.URLParam(r, "id")
		if convoID == "" {
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(
					"Conversation ID is required as a URL parameter",
				))

			return
		}

		photos, err := h.conversationService.GetMatchPhotos(ctx, convoID, userID)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetMatchPhotos", err)
			return
		}

		render.Json(w, http.StatusOK, dto.GetMatchPhotosResponse{
			Photos: mapper.MapPhotosToDTO(photos),
		})
	}
}

func (h *handler) GetConversationScore() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		convoID := chi.URLParam(r, "id")
		if convoID == "" {
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(
					"Conversation ID is required as a URL parameter",
				))

			return
		}

		sharedScore, err := h.conversationService.GetConversationScore(ctx, userID, convoID)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetConversationScore", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetConversationScoreResponse(sharedScore))
	}
}

func (h *handler) GetConversations() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		conversations, err := h.conversationService.GetConversations(ctx, userID)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetConversations", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetConversationsResponse(conversations))
	}
}

func (h *handler) SendMessage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		convoID := chi.URLParam(r, "id")
		if convoID == "" {
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(
					"Conversation ID is required as a URL parameter",
				))

			return
		}

		var req dto.SendMessageRequest
		// Validates and decodes request
		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate send message request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse("client message id and type is required"),
			)

			return
		}

		msg, err := h.conversationService.SendMessage(ctx, nil, mapper.MapSendMessageRequestToDomain(req, convoID, userID))
		if err != nil {
			h.handleServiceErrorResponse(w, r, "SendMessage", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToSendMessageResponse(&msg))
	}
}

func (h *handler) GetConversationMessages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		convoID := chi.URLParam(r, "id")
		if convoID == "" {
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(
					"Conversation ID is required as a URL parameter",
				))

			return
		}

		messages, err := h.conversationService.GetMessages(ctx, convoID, userID)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetConversationMessages", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetConversationMessagesResponse(messages))
	}
}

func (h *handler) handleServiceErrorResponse(w http.ResponseWriter, r *http.Request, handlerName string, err error) {
	if render.ErrorCausedByTimeoutOrClientCancellation(w, r, h.logger, err) {
		return
	}

	h.logger.Sugar().Errorw(fmt.Sprintf("%s failure", handlerName), "error", err.Error())
	statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
	render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, conversation.ErrInvalidMessage):
		return http.StatusBadRequest, "Invalid message"
	case errors.Is(err, conversation.ErrInvalidTextMessage):
		return http.StatusBadRequest, "Invalid text message. Make sure your text is not blank"
	case errors.Is(err, conversation.ErrTextTooLong):
		return http.StatusBadRequest, fmt.Sprintf("Text message is too long. Please keep it under %v characters", constants.MaxTextLengthRunes)
	case errors.Is(err, conversation.ErrMissingRequiredFieldToSendVoicenote):
		return http.StatusBadRequest, "Sending a voice note requires a media_url and media_seconds"
	case errors.Is(err, conversation.ErrVoiceNoteTooLong):
		return http.StatusBadRequest, fmt.Sprintf("Voice note too long. Must be less than %v seconds", constants.MaxVoiceNoteLengthInSeconds)
	case errors.Is(err, commonErrors.ErrInvalidMediaUrl):
		return http.StatusBadRequest, "Invalid media url"
	case errors.Is(err, conversation.ErrGifMessageMissingURL):
		return http.StatusBadRequest, "Invalid gif message. Make sure your gif url is not blank"
	case errors.Is(err, conversation.ErrInvalidMessageType):
		return http.StatusBadRequest, fmt.Sprintf("Invalid message type. Must be one of '%s','%s','%s' or '%s'", domain.MessageTypeVoice, domain.MessageTypeSystem, domain.MessageTypeText, domain.MessageTypeGif)
	case errors.Is(err, convostorage.ErrClientMsgIDNotUnique):
		return http.StatusBadRequest, "Client message ID must be unique. Please generate a new one"
	case errors.Is(err, convostorage.ErrNotConversationParticipant):
		return http.StatusForbidden, "You are not allowed to access this conversation"
	case errors.Is(err, convostorage.ErrNonExistentConversation):
		return http.StatusForbidden, "You are not allowed to access this conversation"
	case errors.Is(err, interaction.ErrInvalidDirection):
		return http.StatusBadRequest, "Invalid direction. Direction must be 'incoming'"
	case errors.Is(err, storage.ErrUserDoesNotExists):
		return http.StatusNotFound, "User does not exist"
	case errors.Is(err, interactionstorage.ErrAlreadySwiped):
		return http.StatusConflict, "You've already swiped on this user"
	case errors.Is(err, conversation.ErrRevealNotEligible):
		return http.StatusForbidden, "You haven't built enough connection to reveal yet"
	case errors.Is(err, conversation.ErrRevealAlreadyInitiated):
		return http.StatusConflict, "Reveal already initiated"
	case errors.Is(err, conversation.ErrRevealRequestExpired):
		return http.StatusGone, "Reveal request has expired"
	case errors.Is(err, conversation.ErrConversationNotRevealed):
		return http.StatusForbidden, "Photos not revealed yet"
	default:
		return http.StatusInternalServerError, commonMessages.InternalServerErrorMsg
	}
}
