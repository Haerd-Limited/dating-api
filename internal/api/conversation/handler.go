package conversation

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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
	Unmatch() http.HandlerFunc
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
			render.HandleServiceErrorResponse(h.logger, w, r, "GetConversations", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetConversationsResponse(conversations, userID))
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
			render.HandleServiceErrorResponse(h.logger, w, r, "SendMessage", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToSendMessageResponse(&msg, userID))
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
			render.HandleServiceErrorResponse(h.logger, w, r, "GetConversationMessages", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetConversationMessagesResponse(messages, userID))
	}
}

func (h *handler) Unmatch() http.HandlerFunc {
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

		var req dto.UnmatchRequest

		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate unmatch request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse("reason is required and cannot be empty"),
			)

			return
		}

		// Validate the request
		if err := req.Validate(); err != nil {
			h.logger.Sugar().Warnw("unmatch request validation failed", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(err.Error()),
			)

			return
		}

		err = h.conversationService.Unmatch(ctx, userID, convoID, strings.TrimSpace(req.Reason))
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "Unmatch", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, dto.UnmatchResponse{
			Message: "Successfully unmatched",
		})
	}
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
	default:
		return http.StatusInternalServerError, commonMessages.InternalServerErrorMsg
	}
}
