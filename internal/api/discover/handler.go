package discover

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/discover/dto"
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
	GetDiscoverWithFilters() http.HandlerFunc
	GetVoiceWorthHearing() http.HandlerFunc
	GetUserPreferences() http.HandlerFunc
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
			render.HandleServiceErrorResponse(h.logger, w, r, "GetDiscoverFeed", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetDiscoverResponse(result))
	}
}

func (h *handler) GetDiscoverWithFilters() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		// Parse request body for filters
		var req dto.GetDiscoverRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("Invalid request body"))
			return
		}

		// Set defaults if not provided
		if req.Limit <= 0 {
			req.Limit = 10
		}

		if req.Offset < 0 {
			req.Offset = 0
		}

		result, err := h.discoverService.GetDiscoverFeedWithFilters(ctx, userID, req.Limit, req.Offset, req.Filters)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "GetDiscoverFeedWithFilters", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
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
			render.HandleServiceErrorResponse(h.logger, w, r, "GetVoiceWorthHearing", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetVoicesWorthHearingResponse(result))
	}
}

func (h *handler) GetUserPreferences() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		preferences, err := h.discoverService.GetUserPreferences(ctx, userID)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "GetUserPreferences", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetUserPreferencesResponse(preferences))
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, storage.ErrUserDoesNotExists):
		return http.StatusNotFound, "User does not exist"
	case errors.Is(err, discover.ErrVoiceWorthHearingSearching):
		return http.StatusOK, messages.VoiceWorthHearingSearchingMsg
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
