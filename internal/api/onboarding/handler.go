package onboarding

import (
	standardcontext "context"
	"errors"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/net/context"

	"github.com/Haerd-Limited/dating-api/internal/api/onboarding/dto"
	"github.com/Haerd-Limited/dating-api/internal/api/onboarding/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/onboarding"
	"github.com/Haerd-Limited/dating-api/internal/user"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/validators"
)

type Handler interface {
	GetStep() http.HandlerFunc
	Register() http.HandlerFunc
	Basics() http.HandlerFunc
	Location() http.HandlerFunc
	Lifestyle() http.HandlerFunc
	Beliefs() http.HandlerFunc
	Background() http.HandlerFunc
	WorkAndEducation() http.HandlerFunc
	Languages() http.HandlerFunc
	Photos() http.HandlerFunc
	Prompts() http.HandlerFunc
}

type handler struct {
	logger            *zap.Logger
	onboardingService onboarding.Service
}

func NewOnboardingHandler(
	logger *zap.Logger,
	onboardingService onboarding.Service,
) Handler {
	return &handler{
		logger:            logger,
		onboardingService: onboardingService,
	}
}

func (h *handler) GetStep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (h *handler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.RegisterRequest

		// Validates and decodes request
		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate register request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse("email, phone_number and first_name are required fields"),
			)

			return
		}

		registerDetails, err := mapper.MapRegisterRequestToDomain(req)
		if err != nil {
			h.logger.Sugar().Errorw("failed to map register request to register input", "error", err)
			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
			render.Json(w, statusCode,
				commonMappers.ToSimpleErrorResponse(errMsg),
			)

			return
		}

		result, err := h.onboardingService.Register(ctx, registerDetails)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error registering user", "error", err)
				statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) Basics() http.HandlerFunc {
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

		var req dto.BasicsRequest

		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate basics request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(messages.AllFieldsRequiredMsg),
			)

			return
		}

		basics, err := mapper.MapBasicRequestToDomain(req, userID)
		if err != nil {
			h.logger.Sugar().Warnw("failed to map basics request to domain", "error", err)
			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
			render.Json(w, statusCode,
				commonMappers.ToSimpleErrorResponse(errMsg),
			)
		}

		result, err := h.onboardingService.Basics(ctx, basics)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error updating basics", "error", err)
				statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) Location() http.HandlerFunc {
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

		var req dto.LocationRequest

		// Validates and decodes request
		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate location request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(messages.AllFieldsRequiredMsg),
			)

			return
		}

		locationDetails, err := mapper.MapLocationRequestToDomain(req, userID)
		if err != nil {
			h.logger.Sugar().Errorw("failed to map location request to domain", "error", err)
			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
			render.Json(w, statusCode,
				commonMappers.ToSimpleErrorResponse(errMsg),
			)

			return
		}

		result, err := h.onboardingService.Location(ctx, locationDetails)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error updating location details", "error", err)
				statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) Lifestyle() http.HandlerFunc {
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

		var req dto.LifestyleRequest

		// Validates and decodes request
		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate lifestyle request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(messages.AllFieldsRequiredMsg),
			)

			return
		}

		lifestyleDetails, err := mapper.MapLifestyleRequestToDomain(req, userID)
		if err != nil {
			h.logger.Sugar().Errorw("failed to map lifestyle request to domain", "error", err)
			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
			render.Json(w, statusCode,
				commonMappers.ToSimpleErrorResponse(errMsg),
			)

			return
		}

		result, err := h.onboardingService.Lifestyle(ctx, lifestyleDetails)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error updating lifestyle details", "error", err)
				statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) Beliefs() http.HandlerFunc {
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

		var req dto.BeliefsRequest

		// Validates and decodes request
		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate beliefs request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(messages.AllFieldsRequiredMsg),
			)

			return
		}

		beliefDetails, err := mapper.MapBeliefsRequestToDomain(req, userID)
		if err != nil {
			h.logger.Sugar().Errorw("failed to map beliefs request to domain", "error", err)
			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
			render.Json(w, statusCode,
				commonMappers.ToSimpleErrorResponse(errMsg),
			)

			return
		}

		result, err := h.onboardingService.Beliefs(ctx, beliefDetails)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error updating beliefs details", "error", err)
				statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) Background() http.HandlerFunc {
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

		var req dto.BackgroundRequest

		// Validates and decodes request
		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate background request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(messages.AllFieldsRequiredMsg),
			)

			return
		}

		backgroundDetails, err := mapper.MapBackgroundRequestToDomain(req, userID)
		if err != nil {
			h.logger.Sugar().Errorw("failed to map background request to domain", "error", err)
			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
			render.Json(w, statusCode,
				commonMappers.ToSimpleErrorResponse(errMsg),
			)

			return
		}

		result, err := h.onboardingService.Background(ctx, backgroundDetails)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error updating background details", "error", err)
				statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) WorkAndEducation() http.HandlerFunc {
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

		var req dto.WorkAndEducation

		// Validates and decodes request
		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate work and education request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg),
			)

			return
		}

		workAndEducationDetails, err := mapper.MapWorkAndEducationRequestToDomain(req, userID)
		if err != nil {
			h.logger.Sugar().Errorw("failed to map work and education request to domain", "error", err)
			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
			render.Json(w, statusCode,
				commonMappers.ToSimpleErrorResponse(errMsg),
			)

			return
		}

		result, err := h.onboardingService.WorkAndEducation(ctx, workAndEducationDetails)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error updating work and education details", "error", err)
				statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) Languages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (h *handler) Photos() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (h *handler) Prompts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

const (
	InvalidUsernameLengthMsg  = "Username must be between 3 and 20 characters long"
	UsernameContainsSpacesMsg = "Username cannot contain spaces"
)

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, standardcontext.Canceled):
		return constants.StatusClientClosedRequest, messages.RequestCancelledMsg
	case errors.Is(err, standardcontext.DeadlineExceeded):
		return http.StatusRequestTimeout, messages.RequestTimeoutMsg
	case errors.Is(err, mapper.ErrInvalidID):
		return http.StatusBadRequest, messages.InvalidIDMsg
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
	case errors.Is(err, mapper.ErrInvalidNameLength):
		return http.StatusBadRequest, InvalidUsernameLengthMsg
	case errors.Is(err, mapper.ErrNameContainsSpaces):
		return http.StatusBadRequest, UsernameContainsSpacesMsg

	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
