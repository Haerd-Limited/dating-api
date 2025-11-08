package notification

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/notification/dto"
	"github.com/Haerd-Limited/dating-api/internal/notification"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	RegisterDeviceToken() http.HandlerFunc
	RemoveDeviceToken() http.HandlerFunc
}

type handler struct {
	logger             *zap.Logger
	notificationSender notification.Service
}

func NewNotificationHandler(logger *zap.Logger, svc notification.Service) Handler {
	return &handler{
		logger:             logger,
		notificationSender: svc,
	}
}

func (h *handler) RegisterDeviceToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.DeviceTokenRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("token is required"))
			return
		}

		if err := h.notificationSender.RegisterDeviceToken(ctx, userID, req.Token); err != nil {
			h.logger.Sugar().Errorw("failed to register device token", "error", err, "userID", userID)
			render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse("failed to store device token"))

			return
		}

		render.Json(w, http.StatusCreated, map[string]string{"status": "ok"})
	}
}

func (h *handler) RemoveDeviceToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.DeviceTokenRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("token is required"))
			return
		}

		if err := h.notificationSender.RemoveDeviceToken(ctx, userID, req.Token); err != nil {
			h.logger.Sugar().Errorw("failed to remove device token", "error", err, "userID", userID)
			render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse("failed to remove device token"))

			return
		}

		render.Json(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
