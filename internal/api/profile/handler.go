package profile

import (
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/profile/dto"
	"github.com/Haerd-Limited/dating-api/internal/api/profile/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/profile"
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
}

type handler struct {
	logger         *zap.Logger
	profileService profile.Service
}

func NewProfileHandler(
	logger *zap.Logger,
	profileService profile.Service,
) Handler {
	return &handler{
		logger:         logger,
		profileService: profileService,
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
			h.handleServiceErrorResponse(w, r, "GetMyProfile", err)
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

		err = h.profileService.UpdateProfile(ctx, model)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "UpdateMyProfile", err)
			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Profile successfully updated"))
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
