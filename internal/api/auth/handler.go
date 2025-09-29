package auth

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/auth/dto"
	"github.com/Haerd-Limited/dating-api/internal/api/auth/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/auth"
	"github.com/Haerd-Limited/dating-api/internal/auth/storage"
	"github.com/Haerd-Limited/dating-api/internal/user"
	userstorage "github.com/Haerd-Limited/dating-api/internal/user/storage"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	VerifyCode() http.HandlerFunc
	RequestCode() http.HandlerFunc
	Refresh() http.HandlerFunc
	Logout() http.HandlerFunc
}

type handler struct {
	logger      *zap.Logger
	authService auth.Service
}

func NewAuthHandler(
	logger *zap.Logger,
	authService auth.Service,
) Handler {
	return &handler{
		logger:      logger,
		authService: authService,
	}
}

const (
	InvalidBodyMsg         = "invalid body"
	InvalidRefreshTokenMsg = "Please provide your refresh token."
)

func (h *handler) RequestCode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.RequestCodeRequest

		// Validates and decodes request
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Errorw("failed to decode and validate request code request body", "context", ctx, "error", err)

			render.Json(
				w,
				http.StatusBadRequest,
				mapper.ToAuthResponse(
					nil,
					InvalidBodyMsg,
				))

			return
		}

		ip := requestIP(r)

		sentTo, err := h.authService.RequestCode(ctx, mapper.MapRequestCodeRequestToDomain(req, ip))
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // don't write a response
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				// Anti-enumeration: still 200 OK, but log server-side
				h.logger.Warn("failed to request code", zap.Error(err))
			}
		}

		render.Json(w, http.StatusOK, mapper.ToRequestCodeResponse(sentTo))
	}
}

func (h *handler) VerifyCode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.VerifyCodeRequest

		// Validates and decodes request
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Errorw("failed to decode and validate verify code request body", "context", ctx, "error", err)

			render.Json(
				w,
				http.StatusBadRequest,
				mapper.ToAuthResponse(
					nil,
					InvalidBodyMsg,
				))

			return
		}

		ip := requestIP(r)

		result, err := h.authService.VerifyCode(ctx, mapper.ToVerifyDomain(req, ip))
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // don't write a response
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			case errors.Is(err, auth.ErrUserAlreadyRegistered) && req.Purpose == "register":
				h.logger.Sugar().Warn("user already registered", zap.Error(err))
				render.Json(
					w,
					http.StatusConflict,
					commonMappers.ToSimpleErrorResponse(
						"A user with this number is already registered. Please try and log in",
					),
				)

				return
			default:
				// Anti-enumeration: still 200 OK, but log server-side
				h.logger.Warn("failed to verify code", zap.Error(err))
			}
		}

		render.Json(w, http.StatusOK, mapper.ToAuthResponse(result, "Code verified successfully"))
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
				mapper.ToAuthResponse(
					nil,
					InvalidRefreshTokenMsg,
				))

			return
		}

		result, err := h.authService.RefreshToken(ctx, mapper.MapRefreshRequestToDomain(req))
		if err != nil {
			switch { // todo: refactor this as it's repeated in all handlers
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // don't write a response
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error refreshing tokens", "error", err)
				code, msg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, code, mapper.ToAuthResponse(
					nil,
					msg,
				))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.ToAuthResponse(result, "Tokens refreshed successfully"))
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
				commonMappers.ToSimpleErrorResponse(
					InvalidRefreshTokenMsg,
				))

			return
		}

		logoutInput := mapper.MapLogoutRequestToDomain(req)

		err := h.authService.RevokeRefreshToken(ctx, logoutInput)
		if err != nil {
			switch { // todo: refactor this as it's repeated in all handlers
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // don't write a response
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error logging out", "error", err)
				code, msg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, code, commonMappers.ToSimpleErrorResponse(
					msg,
				))

				return
			}
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Logged out successfully"))
	}
}

// prefer header, then remote addr; strip port & brackets; take the first XFF hop
func requestIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil { // if it's not host:port, fall back
		host = r.RemoteAddr
	}

	return strings.Trim(host, "[]") // remove IPv6 brackets
}

const (
	InvalidCredentialsMsg    = "Incorrect email or password. Please try again."
	TokenRevokedOrExpiredMsg = "Your refresh token has been revoked or expired"
	InvalidEmailMsg          = "Please enter a valid email address."
	MissingRequiredFieldMsg  = "Please fill in all required fields."
)

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, context.Canceled):
		return constants.StatusClientClosedRequest, messages.RequestCancelledMsg
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusRequestTimeout, messages.RequestTimeoutMsg
	case errors.Is(err, http.ErrNotMultipart):
		return http.StatusBadRequest, messages.InvalidUploadFormMsg
	case errors.Is(err, userstorage.ErrEmailAlreadyExists):
		return http.StatusConflict, messages.EmailAlreadyExistsMsg
	case errors.Is(err, userstorage.ErrUserDetailsAlreadyExists):
		return http.StatusConflict, messages.UserDetailsAlreadyExistsMsg
	case errors.Is(err, user.ErrInvalidCredentials):
		return http.StatusBadRequest, InvalidCredentialsMsg
	case errors.Is(err, auth.ErrRefreshTokenRevoked), errors.Is(err, auth.ErrRefreshTokenExpired), errors.Is(err, storage.ErrRefreshTokenNotFound):
		return http.StatusUnauthorized, TokenRevokedOrExpiredMsg
	case errors.Is(err, auth.ErrRefreshTokenAlreadyRevoked):
		return http.StatusOK, TokenRevokedOrExpiredMsg
	case errors.Is(err, commonErrors.ErrInvalidDOBFormat):
		return http.StatusBadRequest, messages.InvalidDobMsg
	case errors.Is(err, commonErrors.ErrInvalidDob):
		return http.StatusBadRequest, messages.InvalidDobMsg
	case errors.Is(err, commonErrors.ErrInvalidGender):
		return http.StatusBadRequest, messages.InvalidGenderMsg
	case errors.Is(err, commonErrors.ErrInvalidEmail):
		return http.StatusBadRequest, InvalidEmailMsg
	case errors.Is(err, commonErrors.ErrMissingRequiredField):
		return http.StatusBadRequest, MissingRequiredFieldMsg

	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
