package conversation

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/conversation/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/conversation"
	"github.com/Haerd-Limited/dating-api/internal/interaction"
	storage2 "github.com/Haerd-Limited/dating-api/internal/interaction/storage"
	"github.com/Haerd-Limited/dating-api/internal/user/storage"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
)

type Handler interface {
	GetConversations() http.HandlerFunc
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
			authHeader := r.Header.Get("Authorization")
			h.logger.Sugar().Errorw("missing user ID", "authHeader", authHeader)

			render.Json(
				w,
				http.StatusUnauthorized,
				commonMappers.ToSimpleErrorResponse(
					messages.AuthenticationRequiredMsg,
				))

			return
		}

		conversations, err := h.conversationService.GetConversations(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error getting conversations", "error", err)
				statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.MapConversationsToDtos(conversations))
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, interaction.ErrInvalidDirection):
		return http.StatusBadRequest, "Invalid direction. Direction must be 'incoming'"
	case errors.Is(err, storage.ErrUserDoesNotExists):
		return http.StatusNotFound, "User does not exist"
	case errors.Is(err, storage2.ErrAlreadySwiped):
		return http.StatusConflict, "You've already swiped on this user"
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
