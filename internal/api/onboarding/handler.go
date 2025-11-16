package onboarding

import (
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/onboarding/dto"
	"github.com/Haerd-Limited/dating-api/internal/api/onboarding/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/onboarding"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	"github.com/Haerd-Limited/dating-api/internal/user"
	userstorage "github.com/Haerd-Limited/dating-api/internal/user/storage"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	GetStep() http.HandlerFunc
	Intro() http.HandlerFunc
	Basics() http.HandlerFunc
	Stats() http.HandlerFunc
	Location() http.HandlerFunc
	Lifestyle() http.HandlerFunc
	Beliefs() http.HandlerFunc
	Background() http.HandlerFunc
	WorkAndEducation() http.HandlerFunc
	Languages() http.HandlerFunc
	Photos() http.HandlerFunc
	Prompts() http.HandlerFunc
	Profile() http.HandlerFunc
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
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		result, err := h.onboardingService.GetUserCurrentStep(ctx, userID)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "GetStep", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) Stats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		result, err := h.onboardingService.GetPreregistrationStats(ctx)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "Stats", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, result)
	}
}

func (h *handler) Intro() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.IntroRequest

		// Validates and decodes request
		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate intro request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse("email and first_name are required fields"),
			)

			return
		}

		result, err := h.onboardingService.Intro(ctx, mapper.MapIntroRequestToDomain(req, userID))
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "Intro", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) Basics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
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

		result, err := h.onboardingService.Basics(ctx, mapper.MapBasicRequestToDomain(req, userID))
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "Basics", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) Location() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
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

		result, err := h.onboardingService.Location(ctx, mapper.MapLocationRequestToDomain(req, userID))
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "Location", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) Lifestyle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
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

		result, err := h.onboardingService.Lifestyle(ctx, mapper.MapLifestyleRequestToDomain(req, userID))
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "Lifestyle", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) Beliefs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
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

		result, err := h.onboardingService.Beliefs(ctx, mapper.MapBeliefsRequestToDomain(req, userID))
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "Beliefs", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) Background() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
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

		result, err := h.onboardingService.Background(ctx, mapper.MapBackgroundRequestToDomain(req, userID))
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "Background", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) WorkAndEducation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.WorkAndEducationRequest

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

		result, err := h.onboardingService.WorkAndEducation(ctx, mapper.MapWorkAndEducationRequestToDomain(req, userID))
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "WorkAndEducation", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) Languages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.LanguagesRequest

		// Validates and decodes request
		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate languages request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg),
			)

			return
		}

		result, err := h.onboardingService.Languages(ctx, mapper.MapLanguagesRequestToDomain(req, userID))
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "Languages", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) Photos() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.PhotosRequest

		// Validates and decodes request
		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate photos request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(messages.AllFieldsRequiredMsg),
			)

			return
		}

		result, err := h.onboardingService.Photos(ctx, mapper.MapPhotosRequestToDomain(req, userID))
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "Photos", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) Prompts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.PromptsRequest

		// Validates and decodes request
		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate prompts request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(fmt.Sprintf("Please provide at least 4 prompts. Maximum %v prompts. Prompt type, url and position are required fields", constants.MaximumNumberOfPrompts)),
			)

			return
		}

		result, err := h.onboardingService.Prompts(ctx, mapper.MapPromptsRequestToDomain(req, userID))
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "Prompts", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func (h *handler) Profile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.ProfileRequest

		// Validates and decodes request
		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate profile request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg),
			)

			return
		}

		result, err := h.onboardingService.Profile(ctx, mapper.MapProfileToDomain(req, userID))
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "Profile", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.ToOnboardingResponse(result))
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, profile.ErrNotEnoughPhotosProvided):
		return http.StatusBadRequest, "Please provide exactly 6 photos"
	case errors.Is(err, profile.ErrContainsSocialMediaPromotion):
		return http.StatusBadRequest, messages.SocialsNotAllowedMsg
	case errors.Is(err, profile.ErrInvalidBirthdate):
		return http.StatusBadRequest, fmt.Sprintf("Invalid birthdate. You must be at least %v and at most %v", constants.MinAge, constants.MaxAge)
	case errors.Is(err, profile.ErrInvalidPromptPosition):
		return http.StatusBadRequest, "Invalid prompt position"
	case errors.Is(err, profile.ErrDuplicatePromptPosition):
		return http.StatusBadRequest, "Duplicate prompt position"
	case errors.Is(err, commonErrors.ErrInvalidEmail):
		return http.StatusBadRequest, messages.InvalidEmailMsg
	case errors.Is(err, onboarding.ErrIncorrectStepCalled):
		return http.StatusBadRequest, "Incorrect step called"
	case errors.Is(err, profile.ErrInvalidID):
		return http.StatusBadRequest, messages.InvalidIDMsg
	case errors.Is(err, http.ErrNotMultipart):
		return http.StatusBadRequest, messages.InvalidUploadFormMsg
	case errors.Is(err, userstorage.ErrEmailAlreadyExists):
		return http.StatusConflict, messages.EmailAlreadyExistsMsg
	case errors.Is(err, userstorage.ErrUserDetailsAlreadyExists):
		return http.StatusConflict, messages.UserDetailsAlreadyExistsMsg
	case errors.Is(err, commonErrors.ErrInvalidDOBFormat):
		return http.StatusBadRequest, messages.InvalidDobMsg
	case errors.Is(err, commonErrors.ErrInvalidDob):
		return http.StatusBadRequest, messages.InvalidDobMsg
	case errors.Is(err, commonErrors.ErrInvalidGender):
		return http.StatusBadRequest, messages.InvalidGenderMsg
	case errors.Is(err, user.ErrInvalidNameLength):
		return http.StatusBadRequest, "Username must be between 3 and 20 characters long"
	case errors.Is(err, user.ErrNameContainsSpaces):
		return http.StatusBadRequest, "Username cannot contain spaces"
	case errors.Is(err, onboarding.ErrPreregistrationCapped):
		return http.StatusConflict, "Registration for this cohort is full. Please check back later."

	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
