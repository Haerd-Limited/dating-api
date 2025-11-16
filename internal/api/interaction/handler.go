package interaction

import (
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/interaction/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/api/profile/dto"
	convostorage "github.com/Haerd-Limited/dating-api/internal/conversation/storage"
	"github.com/Haerd-Limited/dating-api/internal/interaction"
	storage2 "github.com/Haerd-Limited/dating-api/internal/interaction/storage"
	"github.com/Haerd-Limited/dating-api/internal/user/storage"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	Create() http.HandlerFunc
	GetLikes() http.HandlerFunc
}

type handler struct {
	logger             *zap.Logger
	interactionService interaction.Service
}

func NewInteractionHandler(
	logger *zap.Logger,
	interactionService interaction.Service,
) Handler {
	return &handler{
		logger:             logger,
		interactionService: interactionService,
	}
}

func (h *handler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.SwipesRequest

		// Validates and decodes request
		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate swipes request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(
					fmt.Sprintf("target_user_id and action are required and action field must be one of '%s','%s' or '%s'. message_type is optional and must be one of '%s' or '%s' if provided",
						constants.ActionLike,
						constants.ActionPass,
						constants.ActionSuperlike,
						constants.MessageTypeText,
						constants.MessageTypeVoice,
					)),
			)

			return
		}

		result, err := h.interactionService.CreateSwipe(ctx, mapper.SwipesRequestToDomain(req, userID))
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "CreateSwipe", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusCreated, mapper.MapToSwipesResponse(result))
	}
}

func (h *handler) GetLikes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		dir := r.URL.Query().Get("direction")
		limit := request.ParseQueryInt(r, "limit", 5)
		offset := request.ParseQueryInt(r, "offset", 0)

		profiles, err := h.interactionService.GetLikes(ctx, userID, dir, offset, limit)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "GetLikes", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetLikesResponse(&profiles))
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, convostorage.ErrClientMsgIDNotUnique):
		return http.StatusBadRequest, "Client message ID must be unique"
	case errors.Is(err, interaction.ErrPromptIDRequiredToLikeUser):
		return http.StatusBadRequest, "Prompt ID is required to like a profile"
	case errors.Is(err, interaction.ErrSelfLike):
		return http.StatusBadRequest, "As much as we promote self love, you cannot like yourself"
	case errors.Is(err, interaction.ErrInvalidAction):
		return http.StatusBadRequest, fmt.Sprintf("Invalid action. Action must be '%s','%s' or '%s'", constants.ActionLike, constants.ActionPass, constants.ActionSuperlike)
	case errors.Is(err, interaction.ErrMissingRequiredFieldsForLikeWithMessage):
		return http.StatusBadRequest, "Sending a like with a message also requires message_type, message,prompt_id and a generated client_msg_id"
	case errors.Is(err, interaction.ErrLikedAVhwUser):
		return http.StatusBadRequest, "You can only superlike or pass a Voices Worth Hearing user"
	case errors.Is(err, interaction.ErrWeeklySuperlikeLimitReached):
		return http.StatusForbidden, "You've already used your superlike this week"
	case errors.Is(err, interaction.ErrInvalidDirection):
		return http.StatusBadRequest, "Invalid direction. Direction must be 'incoming'"
	case errors.Is(err, storage.ErrUserDoesNotExists):
		return http.StatusNotFound, "User does not exist"
	case errors.Is(err, commonErrors.ErrInvalidDOBFormat):
		return http.StatusBadRequest, messages.InvalidDobMsg
	case errors.Is(err, storage2.ErrAlreadySwiped):
		return http.StatusConflict, "You've already swiped on this user"
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
