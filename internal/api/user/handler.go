package user

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/user/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/user"
	"github.com/Haerd-Limited/dating-api/internal/user/domain"
	"github.com/Haerd-Limited/dating-api/internal/user/storage"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/validators"
)

type Handler interface {
	ViewProfile() http.HandlerFunc
	MyProfile() http.HandlerFunc
	UpdateProfile() http.HandlerFunc
}

type handler struct {
	logger      *zap.Logger
	userService user.Service
}

func NewUserHandler(
	logger *zap.Logger,
	userService user.Service,
) Handler {
	return &handler{
		logger:      logger,
		userService: userService,
	}
}
func (h *handler) ViewProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := chi.URLParam(r, "userID")

		viewerID, ok := context.UserIDFromContext(ctx)
		if !ok {
			authHeader := r.Header.Get("Authorization")
			h.logger.Sugar().Errorw("missing user ID", "authHeader", authHeader)

			render.Json(
				w,
				http.StatusUnauthorized,
				mapper.ToGetProfileResponse(
					nil,
					messages.AuthenticationRequiredMsg,
				))

			return
		}

		result, err := h.userService.GetUserProfile(
			ctx,
			&domain.ViewProfile{
				ViewerID:     viewerID,
				TargetUserID: userID,
			})
		if err != nil {
			h.logger.Error("failed to get user profile", zap.Any("context", ctx), zap.Error(err))

			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
			render.Json(
				w,
				statusCode,
				mapper.ToGetProfileResponse(nil, errMsg))

			return
		}

		render.Json(w, http.StatusOK, mapper.ToGetProfileResponse(result, "User profile retrieved successfully"))
	}
}

func (h *handler) MyProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := context.UserIDFromContext(ctx)
		if !ok {
			authHeader := r.Header.Get("Authorization")
			h.logger.Sugar().Errorw("missing user ID", "authHeader", authHeader)

			render.Json(
				w,
				http.StatusUnauthorized,
				mapper.ToGetMyProfileResponse(
					nil,
					messages.AuthenticationRequiredMsg,
				))

			return
		}

		profile, err := h.userService.GetUserProfile(
			ctx,
			&domain.ViewProfile{
				ViewerID:     userID,
				TargetUserID: userID,
			})
		if err != nil {
			h.logger.Error("failed to get own profile", zap.Any("context", ctx), zap.Error(err))
			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)

			render.Json(
				w,
				statusCode,
				mapper.ToGetMyProfileResponse(nil, errMsg))

			return
		}

		render.Json(w, http.StatusOK, mapper.ToGetMyProfileResponse(profile, "User profile retrieved successfully"))
	}
}

func (h *handler) UpdateProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := context.UserIDFromContext(ctx)
		if !ok {
			authHeader := r.Header.Get("Authorization")
			h.logger.Sugar().Errorw("missing user ID", "authHeader", authHeader)

			render.Json(
				w,
				http.StatusUnauthorized,
				mapper.ToGetMyProfileResponse(
					nil,
					messages.AuthenticationRequiredMsg,
				))

			return
		}

		// Validates and decodes request
		request, err := validators.DecodeAndValidateUpdateProfileForm(r)
		if err != nil {
			h.logger.Sugar().Errorw("failed to decode and validate register request body", "context", ctx, "error", err)

			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)

			render.Json(
				w,
				statusCode,
				mapper.ToUpdateProfileResponse(
					nil,
					errMsg,
				))

			return
		}

		input, err := mapper.UpdateProfileRequestToDomain(request, userID)
		if err != nil {
			h.logger.Sugar().Errorw("failed to map request to update profile input", "context", ctx, "error", err)

			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)

			render.Json(w, statusCode,
				mapper.ToUpdateProfileResponse(
					nil,
					errMsg,
				),
			)

			return
		}

		profile, err := h.userService.UpdateUserProfile(ctx, input)
		if err != nil {
			h.logger.Error("failed to update your profile", zap.Any("context", ctx), zap.Error(err))

			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)

			render.Json(
				w,
				statusCode,
				mapper.ToUpdateProfileResponse(nil, errMsg))

			return
		}

		render.Json(w, http.StatusOK, mapper.ToUpdateProfileResponse(profile, "Your profile has been updated successfully"))
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, storage.ErrUserDoesNotExists):
		return http.StatusNotFound, "User does not exist"
	case errors.Is(err, user.ErrEmailAlreadyExists):
		return http.StatusConflict, messages.EmailAlreadyExistsMsg
	case errors.Is(err, user.ErrUserNameAlreadyExists):
		return http.StatusConflict, messages.UserNameAlreadyExistsMsg
	case errors.Is(err, user.ErrUserDetailsAlreadyExists):
		return http.StatusConflict, messages.UserDetailsAlreadyExistsMsg
	case errors.Is(err, commonErrors.ErrInvalidDob):
		return http.StatusBadRequest, messages.InvalidDobMsg
	case errors.Is(err, commonErrors.ErrInvalidGender):
		return http.StatusBadRequest, messages.InvalidGenderMsg
	case errors.Is(err, http.ErrNotMultipart):
		return http.StatusBadRequest, messages.InvalidUploadFormMsg
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
