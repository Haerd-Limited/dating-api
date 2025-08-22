package onboarding

import (
	standardcontext "context"
	"errors"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/service"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/onboarding/dto"
	"github.com/Haerd-Limited/dating-api/internal/api/onboarding/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/user"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/validators"
)

type Handler interface {
	Patch() http.HandlerFunc
	PatchVisibility() http.HandlerFunc
	Complete() http.HandlerFunc
	State() http.HandlerFunc
}

type handler struct {
	logger            *zap.Logger
	onboardingService service.Service
}

func NewOnboardingHandler(
	logger *zap.Logger,
	onboardingService service.Service,
) Handler {
	return &handler{
		logger:            logger,
		onboardingService: onboardingService,
	}
}

func (h *handler) Patch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := context.UserIDFromContext(ctx)
		if !ok {
			authHeader := r.Header.Get("Authorization")
			h.logger.Sugar().Errorw("missing user ID", "authHeader", authHeader)

			render.Json(
				w,
				http.StatusUnauthorized,
				commonMappers.ToSimpleMessageResponse(
					messages.AuthenticationRequiredMsg,
				))

			return
		}

		var req dto.UpdateOnboardingRequest
		// Validates and decodes request
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Errorw("failed to decode and validate onboarding request body", "context", ctx, "error", err)

			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleMessageResponse(
					"invalid request body",
				))

			return
		}

		progress, err := h.onboardingService.Patch(ctx, mapper.ToDomain(userID, req))
		if err != nil {
			switch {
			case errors.Is(err, standardcontext.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // don't write a response
			case errors.Is(err, standardcontext.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleMessageResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("failed to patch onboarding", "context", ctx, "error", err)
				code, msg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, code, commonMappers.ToSimpleMessageResponse(msg))
				return
			}
		}

		render.Json(
			w,
			http.StatusOK,
			mapper.ToOnboardingProgressResponse(progress),
		)
	}
}

func (h *handler) PatchVisibility() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (h *handler) Complete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (h *handler) State() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, standardcontext.Canceled):
		return constants.StatusClientClosedRequest, messages.RequestCancelledMsg
	case errors.Is(err, standardcontext.DeadlineExceeded):
		return http.StatusRequestTimeout, messages.RequestTimeoutMsg
	case errors.Is(err, http.ErrNotMultipart):
		return http.StatusBadRequest, messages.InvalidUploadFormMsg
	case errors.Is(err, user.ErrEmailAlreadyExists):
		return http.StatusConflict, messages.EmailAlreadyExistsMsg
	case errors.Is(err, user.ErrUserDetailsAlreadyExists):
		return http.StatusConflict, messages.UserDetailsAlreadyExistsMsg
	case errors.Is(err, validators.ErrInvalidDOBFormat):
		return http.StatusBadRequest, messages.InvalidDobMsg
	case errors.Is(err, commonErrors.ErrInvalidDob):
		return http.StatusBadRequest, messages.InvalidDobMsg
	case errors.Is(err, commonErrors.ErrInvalidGender):
		return http.StatusBadRequest, messages.InvalidGenderMsg
	case errors.Is(err, service.ErrInvalidAgePreference):
		return http.StatusBadRequest, "Invalid age preference"

	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
