package user

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/user/dto"
	"github.com/Haerd-Limited/dating-api/internal/api/user/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/discover"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	"github.com/Haerd-Limited/dating-api/internal/user"
	"github.com/Haerd-Limited/dating-api/internal/user/storage"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	GetMyProfile() http.HandlerFunc
	UpdateMyProfile() http.HandlerFunc
	GetDiscover() http.HandlerFunc
}

type handler struct {
	logger          *zap.Logger
	userService     user.Service
	profileService  profile.Service
	discoverService discover.Service
}

func NewUserHandler(
	logger *zap.Logger,
	userService user.Service,
	profileService profile.Service,
	discoverService discover.Service,
) Handler {
	return &handler{
		logger:          logger,
		userService:     userService,
		profileService:  profileService,
		discoverService: discoverService,
	}
}

func (h *handler) GetMyProfile() http.HandlerFunc {
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

func (h *handler) UpdateMyProfile() http.HandlerFunc {
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

		var req dto.UpdateProfileRequest

		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate update profile request body", "error", err)
			render.Json(
				w,
				http.StatusInternalServerError,
				commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg),
			)

			return
		}

		model, err := mapper.UpdateProfileRequestToDomain(req, userID)
		if err != nil {
			h.logger.Sugar().Errorw("failed to map update profile request to domain", "error", err)
			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
			render.Json(w, statusCode,
				commonMappers.ToSimpleErrorResponse(errMsg),
			)
		}

		updatedProfile, err := h.profileService.UpdateMyProfile(ctx, model)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error updating user profile", "error", err)
				statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.ProfileToDto(updatedProfile))
	}
}

func (h *handler) GetDiscover() http.HandlerFunc {
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

		limit := request.ParseQueryInt(r, "limit", 10)
		offset := request.ParseQueryInt(r, "offset", 0)

		result, err := h.discoverService.GetDiscoverFeed(ctx, userID, limit, offset)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error getting user discover feed", "error", err)
				statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.FeedProfilesToDto(result))
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, storage.ErrUserDoesNotExists):
		return http.StatusNotFound, "User does not exist"
	case errors.Is(err, commonErrors.ErrInvalidDOBFormat):
		return http.StatusBadRequest, messages.InvalidDobMsg
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
