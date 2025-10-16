package discover

import (
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/discover/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/discover"
	"github.com/Haerd-Limited/dating-api/internal/user/storage"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	GetDiscover() http.HandlerFunc
	GetVoiceWorthHearing() http.HandlerFunc
}

type handler struct {
	logger          *zap.Logger
	discoverService discover.Service
}

func NewDiscoverHandler(
	logger *zap.Logger,
	discoverService discover.Service,
) Handler {
	return &handler{
		logger:          logger,
		discoverService: discoverService,
	}
}

func (h *handler) GetDiscover() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		limit := request.ParseQueryInt(r, "limit", 10)
		offset := request.ParseQueryInt(r, "offset", 0)

		result, err := h.discoverService.GetDiscoverFeed(ctx, userID, limit, offset)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetDiscoverFeed", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetDiscoverResponse(result))
	}
}

func (h *handler) GetVoiceWorthHearing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		result, err := h.discoverService.GetVoiceWorthHearing(ctx, userID)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetVoiceWorthHearing", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetVoicesWorthHearingResponse(result))
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, storage.ErrUserDoesNotExists):
		return http.StatusNotFound, "User does not exist"
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}

func (h *handler) handleServiceErrorResponse(w http.ResponseWriter, r *http.Request, handlerName string, err error) {
	if render.ErrorCausedByTimeoutOrClientCancellation(w, r, h.logger, err) {
		return
	}

	h.logger.Sugar().Errorw(fmt.Sprintf("%s failure", handlerName), "error", err.Error())
	statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
	render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))
}
