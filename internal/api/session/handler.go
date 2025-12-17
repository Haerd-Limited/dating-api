package session

import (
	"net/http"

	"go.uber.org/zap"

	dto "github.com/Haerd-Limited/dating-api/internal/api/session/dto"
	internalsession "github.com/Haerd-Limited/dating-api/internal/session"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
)

type Handler interface {
	TrackAppOpen() http.HandlerFunc
}

type handler struct {
	logger         *zap.Logger
	sessionService internalsession.Service
}

func NewSessionHandler(logger *zap.Logger, sessionService internalsession.Service) Handler {
	return &handler{
		logger:         logger,
		sessionService: sessionService,
	}
}

func (h *handler) TrackAppOpen() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract user ID from context (already authenticated via middleware)
		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		// Track the app open event
		err := h.sessionService.TrackAppOpen(ctx, userID)
		if err != nil {
			h.logger.Sugar().Warnf("failed to track app open: %s", err.Error())
			render.Json(w, http.StatusInternalServerError, dto.TrackAppOpenResponse{
				Message: "Failed to track session",
			})

			return
		}

		// Return success response
		render.Json(w, http.StatusOK, dto.TrackAppOpenResponse{
			Message: "Session tracked successfully",
		})
	}
}
