package auth

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/auth/dto"
	"github.com/Haerd-Limited/dating-api/internal/api/auth/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/auth"
	"github.com/Haerd-Limited/dating-api/internal/user"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/validators"
)

type Handler interface {
	Register() http.HandlerFunc
	Login() http.HandlerFunc
	Refresh() http.HandlerFunc
	Logout() http.HandlerFunc
}

type handler struct {
	logger      *zap.Logger
	authService auth.AuthService
}

func NewAuthHandler(
	logger *zap.Logger,
	authService auth.AuthService,
) Handler {
	return &handler{
		logger:      logger,
		authService: authService,
	}
}

const (
	InvalidLoginInputMsg   = "Please provide both your account's email and password."
	InvalidRefreshTokenMsg = "Please provide your refresh token."
)

func (h *handler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Validates and decodes request
		req, err := validators.DecodeAndValidateRegisterForm(r)
		if err != nil {
			h.logger.Sugar().Errorw("failed to decode and validate register request body", "context", ctx, "error", err)

			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)

			render.Json(
				w,
				statusCode,
				mapper.ToRegisterResponse(
					nil,
					errMsg,
				))

			return
		}

		registerDetails, err := mapper.MapRegisterRequestToDomain(req)
		if err != nil {
			h.logger.Sugar().Errorw("failed to map register request to register input", "error", err)

			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)

			render.Json(w, statusCode,
				mapper.ToRegisterResponse(
					nil,
					errMsg,
				),
			)

			return
		}

		result, err := h.authService.Register(ctx, registerDetails)
		if err != nil {
			h.logger.Sugar().Errorw("Error registering user", "error", err)

			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)

			render.Json(w, statusCode,
				mapper.ToRegisterResponse(
					nil,
					errMsg,
				),
			)

			return
		}

		render.Json(w, http.StatusCreated, mapper.ToRegisterResponse(result, "Registration successful"))
	}
}

func (h *handler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.LoginRequest

		// Validates and decodes request
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Errorw("failed to decode and validate login request body", "context", ctx, "error", err)

			render.Json(
				w,
				http.StatusBadRequest,
				mapper.MapAuthTokensAndUserResponse(
					nil,
					InvalidLoginInputMsg,
				))

			return
		}

		userCredentials, err := h.authService.Login(ctx, mapper.MapLoginRequestToDomain(req))
		if err != nil {
			h.logger.Sugar().Errorw("Error logging in user", "error", err)

			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)

			render.Json(w, statusCode,
				mapper.MapAuthTokensAndUserResponse(
					nil,
					errMsg,
				),
			)

			return
		}

		render.Json(w, http.StatusOK, mapper.MapAuthTokensAndUserResponse(userCredentials, "Login successful"))
	}
}

func (h *handler) Refresh() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.RefreshRequest

		// Validates and decodes request
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Errorw("failed to decode and validate refresh request body", "context", ctx, "error", err)

			render.Json(
				w,
				http.StatusBadRequest,
				mapper.MapAuthTokensToResponse(
					nil,
					InvalidRefreshTokenMsg,
				))

			return
		}

		result, err := h.authService.RefreshToken(ctx, mapper.MapRefreshRequestToDomain(req))
		if err != nil {
			h.logger.Sugar().Errorw("Error refreshing tokens", "context", ctx, "error", err)

			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)

			render.Json(w, statusCode,
				mapper.MapAuthTokensToResponse(
					nil,
					errMsg,
				),
			)

			return
		}

		resp := mapper.MapAuthTokensToResponse(result, "Tokens refreshed successfully")

		render.Json(w, http.StatusOK, resp)
	}
}

func (h *handler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.LogoutRequest

		// Validates and decodes request
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Errorw("failed to decode and validate logout request body", "context", ctx, "error", err)

			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleMessageResponse(
					InvalidRefreshTokenMsg,
				))

			return
		}

		logoutInput := mapper.MapLogoutRequestToDomain(req)

		err := h.authService.RevokeRefreshToken(ctx, logoutInput)
		if err != nil {
			h.logger.Sugar().Errorw("Error logging out", "error", err)

			statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)

			render.Json(w, statusCode,
				commonMappers.ToSimpleMessageResponse(
					errMsg,
				),
			)

			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Logged out successfully"))
	}
}

const (
	InvalidCredentialsMsg     = "Incorrect email or password. Please try again."
	TokenRevokedOrExpiredMsg  = "Your refresh token has been revoked or expired"
	InvalidEmailMsg           = "Please enter a valid email address."
	MissingRequiredFieldMsg   = "Please fill in all required fields."
	InvalidUsernameLengthMsg  = "Username must be between 3 and 20 characters long"
	UsernameContainsSpacesMsg = "Username cannot contain spaces"
)

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, http.ErrNotMultipart):
		return http.StatusBadRequest, messages.InvalidUploadFormMsg
	case errors.Is(err, user.ErrEmailAlreadyExists):
		return http.StatusConflict, messages.EmailAlreadyExistsMsg
	case errors.Is(err, user.ErrUserNameAlreadyExists):
		return http.StatusConflict, messages.UserNameAlreadyExistsMsg
	case errors.Is(err, user.ErrUserDetailsAlreadyExists):
		return http.StatusConflict, messages.UserDetailsAlreadyExistsMsg
	case errors.Is(err, user.ErrInvalidCredentials):
		return http.StatusBadRequest, InvalidCredentialsMsg
	case errors.Is(err, auth.ErrRefreshTokenRevoked), errors.Is(err, auth.ErrRefreshTokenExpired), errors.Is(err, auth.ErrRefreshTokenNotFound):
		return http.StatusUnauthorized, TokenRevokedOrExpiredMsg
	case errors.Is(err, auth.ErrRefreshTokenAlreadyRevoked):
		return http.StatusOK, TokenRevokedOrExpiredMsg
	case errors.Is(err, validators.ErrInvalidDOBFormat):
		return http.StatusBadRequest, messages.InvalidDobMsg
	case errors.Is(err, commonErrors.ErrInvalidDob):
		return http.StatusBadRequest, messages.InvalidDobMsg
	case errors.Is(err, commonErrors.ErrInvalidGender):
		return http.StatusBadRequest, messages.InvalidGenderMsg
	case errors.Is(err, validators.ErrInvalidEmail):
		return http.StatusBadRequest, InvalidEmailMsg
	case errors.Is(err, validators.ErrMissingRequiredField):
		return http.StatusBadRequest, MissingRequiredFieldMsg
	case errors.Is(err, mapper.ErrInvalidUserameLength):
		return http.StatusBadRequest, InvalidUsernameLengthMsg
	case errors.Is(err, mapper.ErrUsernameContainsSpaces):
		return http.StatusBadRequest, UsernameContainsSpacesMsg

	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
