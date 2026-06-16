package broadcast

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	dto "github.com/Haerd-Limited/dating-api/internal/api/broadcast/dto"
	dtoMapper "github.com/Haerd-Limited/dating-api/internal/api/broadcast/dto/mapper"
	internalbroadcast "github.com/Haerd-Limited/dating-api/internal/broadcast"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	ListWaitlistUsers() http.HandlerFunc
	SendBroadcast() http.HandlerFunc
}

type handler struct {
	logger           *zap.Logger
	broadcastService internalbroadcast.Service
}

func NewHandler(logger *zap.Logger, broadcastService internalbroadcast.Service) Handler {
	return &handler{
		logger:           logger,
		broadcastService: broadcastService,
	}
}

func (h *handler) ListWaitlistUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		users, err := h.broadcastService.ListWaitlistUsers(ctx)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "ListWaitlistUsers", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, dtoMapper.WaitlistUsersToResponse(users))
	}
}

func (h *handler) SendBroadcast() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.SendBroadcastRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Warnf("failed to decode and validate broadcast request body: %s", err.Error())
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))

			return
		}

		result, err := h.broadcastService.SendBroadcast(ctx, req.UserIDs, req.Message)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "SendBroadcast", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, dtoMapper.BroadcastResultToResponse(result))
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, internalbroadcast.ErrEmptyMessage):
		return http.StatusBadRequest, "Message is required"
	case errors.Is(err, internalbroadcast.ErrMessageTooLong):
		return http.StatusBadRequest, "Message exceeds maximum length"
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
