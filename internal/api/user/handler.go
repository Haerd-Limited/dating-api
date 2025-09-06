package user

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/user/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	"github.com/Haerd-Limited/dating-api/internal/user"
	"github.com/Haerd-Limited/dating-api/internal/user/storage"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
)

type Handler interface {
	MyProfile() http.HandlerFunc
}

type handler struct {
	logger         *zap.Logger
	userService    user.Service
	profileService profile.Service
}

func NewUserHandler(
	logger *zap.Logger,
	userService user.Service,
	profileService profile.Service,
) Handler {
	return &handler{
		logger:         logger,
		userService:    userService,
		profileService: profileService,
	}
}

func (h *handler) MyProfile() http.HandlerFunc {
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

		userProfile, err := h.profileService.GetMyProfile(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error getting user profile", "error", err)
				statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.ProfileToDto(userProfile))
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, storage.ErrUserDoesNotExists):
		return http.StatusNotFound, "User does not exist"

	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
