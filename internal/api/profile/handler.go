package profile

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/profile/dto"
	"github.com/Haerd-Limited/dating-api/internal/api/profile/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	"github.com/Haerd-Limited/dating-api/internal/user"
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
	GetMyProfile() http.HandlerFunc
	UpdateMyProfile() http.HandlerFunc
	Verify() http.HandlerFunc
	GetVoicePromptTranscript() http.HandlerFunc
	DeleteAccount() http.HandlerFunc
}

type handler struct {
	logger         *zap.Logger
	profileService profile.Service
	userService    user.Service
}

func NewProfileHandler(
	logger *zap.Logger,
	profileService profile.Service,
	userService user.Service,
) Handler {
	return &handler{
		logger:         logger,
		profileService: profileService,
		userService:    userService,
	}
}

func (h *handler) Verify() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		err := h.profileService.VerifyProfile(ctx, userID)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "Verify", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Profile successfully verified"))
	}
}

func (h *handler) GetMyProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		userProfile, err := h.profileService.GetEnrichedProfile(ctx, userID)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "GetMyProfile", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		//todo:update to use a toresponse mapper
		render.Json(w, http.StatusOK, mapper.ProfileToDto(userProfile))
	}
}

func (h *handler) UpdateMyProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.UpdateProfileRequest

		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate update profile request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(messages.InvalidRequestBodyMsg),
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

		err = h.profileService.UpdateProfile(ctx, model)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "UpdateMyProfile", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Profile successfully updated"))
	}
}

func (h *handler) GetVoicePromptTranscript() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		_, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		// Extract voice prompt ID from URL
		idStr := chi.URLParam(r, "id")

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			render.Json(w, http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse("Invalid voice prompt ID"))
			return
		}

		transcript, err := h.profileService.GetTranscript(ctx, id)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "GetVoicePromptTranscript", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, map[string]string{
			"transcript": transcript,
		})
	}
}

func (h *handler) DeleteAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		err := h.userService.DeleteAccount(ctx, userID)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "DeleteAccount", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Account successfully deleted"))
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return http.StatusNotFound, "Voice prompt not found"
	case errors.Is(err, profile.ErrInvalidPromptPosition):
		return http.StatusBadRequest, "Invalid prompt position"
	case errors.Is(err, profile.ErrDuplicatePromptPosition):
		return http.StatusBadRequest, "Duplicate prompt position"
	case errors.Is(err, profile.ErrDuplicatePhotoPosition):
		return http.StatusBadRequest, "Duplicate photo position"
	case errors.Is(err, commonErrors.ErrInvalidMediaUrl):
		return http.StatusBadRequest, "Invalid media url"
	case errors.Is(err, profile.ErrInvalidHeight):
		return http.StatusBadRequest, "Please provide a realistic height"
	case errors.Is(err, profile.ErrInvalidBirthdate):
		return http.StatusBadRequest, fmt.Sprintf("Invalid birthdate. You must be at least %v and at most %v", constants.MinAge, constants.MaxAge)
	case errors.Is(err, profile.ErrContainsSocialMediaPromotion):
		return http.StatusBadRequest, messages.SocialsNotAllowedMsg
	case errors.Is(err, storage.ErrUserDoesNotExists):
		return http.StatusNotFound, "User does not exist"
	case errors.Is(err, commonErrors.ErrInvalidDOBFormat):
		return http.StatusBadRequest, messages.InvalidDobMsg
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
