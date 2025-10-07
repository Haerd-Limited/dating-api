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
	convostorage "github.com/Haerd-Limited/dating-api/internal/conversation/storage"
	"github.com/Haerd-Limited/dating-api/internal/interaction"
	interactionstorage "github.com/Haerd-Limited/dating-api/internal/interaction/storage"
	"github.com/Haerd-Limited/dating-api/internal/user/storage"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
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

func (h *handler) GetConversationScore() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		msg, err := h.conversationService.SendMessage(ctx, mapper.MapSendMessageRequestToDomain(req, convoID, userID))
		if err != nil {
			h.handleServiceErrorResponse(w, r, "SendMessage", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapMessageToDto(&msg)) // todo: update to response
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

	h.logger.Sugar().Errorw(fmt.Sprintf("%s failure", handlerName), "error", err)
	statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
	render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
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
