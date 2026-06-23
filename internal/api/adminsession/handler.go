package adminsession

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	internaladminsession "github.com/Haerd-Limited/dating-api/internal/adminsession"
	"github.com/Haerd-Limited/dating-api/internal/adminsession/domain"
	"github.com/Haerd-Limited/dating-api/internal/api/adminsession/dto"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	CreateSession() http.HandlerFunc
	DeleteSession() http.HandlerFunc
	GetRoster() http.HandlerFunc
}

type handler struct {
	logger         *zap.Logger
	sessionService internaladminsession.Service
}

func NewHandler(logger *zap.Logger, sessionService internaladminsession.Service) Handler {
	return &handler{
		logger:         logger,
		sessionService: sessionService,
	}
}

func (h *handler) CreateSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.CreateSessionRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))
			return
		}

		ip := request.ClientIP(r)

		var ipPtr *string

		if ip != "" {
			ipPtr = &ip
		}

		apiKeyFP := internaladminsession.TokenFingerprint(r.Header.Get("X-Admin-Token"))

		result, err := h.sessionService.CreateSession(ctx, domain.CreateSessionRequest{
			DisplayName: req.DisplayName,
			APIKeyFP:    apiKeyFP,
			IP:          ipPtr,
		})
		if err != nil {
			if errors.Is(err, internaladminsession.ErrInvalidDisplayName) {
				render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid display name"))
				return
			}

			render.HandleServiceErrorResponse(h.logger, w, r, "CreateAdminSession", err, nil)

			return
		}

		render.Json(w, http.StatusOK, dto.CreateSessionResponse{
			SessionToken: result.SessionToken,
			DisplayName:  result.DisplayName,
			ExpiresAt:    result.ExpiresAt.Format(timeRFC3339),
		})
	}
}

func (h *handler) DeleteSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		token := r.Header.Get("X-Admin-Session")
		if token == "" {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("session token required"))
			return
		}

		if err := h.sessionService.DeleteSession(ctx, token); err != nil {
			if errors.Is(err, internaladminsession.ErrSessionNotFound) {
				render.Json(w, http.StatusNotFound, commonMappers.ToSimpleErrorResponse("session not found"))
				return
			}

			render.HandleServiceErrorResponse(h.logger, w, r, "DeleteAdminSession", err, nil)

			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("logged out"))
	}
}

func (h *handler) GetRoster() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.Json(w, http.StatusOK, dto.RosterResponse{
			Names: h.sessionService.Roster(),
		})
	}
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"
