package lookup

import (
	"context"
	"errors"
	"github.com/Haerd-Limited/dating-api/internal/api/lookup/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/lookup"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"go.uber.org/zap"
	"net/http"
)

type Handler interface {
	GetPrompts() http.HandlerFunc
	GetLanguages() http.HandlerFunc
	GetReligions() http.HandlerFunc
	GetPoliticalBeliefs() http.HandlerFunc
}

type handler struct {
	logger        *zap.Logger
	lookupService lookup.Service
}

func NewLookupHandler(
	logger *zap.Logger,
	lookupService lookup.Service,
) Handler {
	return &handler{
		logger:        logger,
		lookupService: lookupService,
	}
}

func (h *handler) GetPrompts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		prompts, err := h.lookupService.GetPrompts(ctx)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error getting prompts", "error", err)
				render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.MapToGetPromptsResponse(prompts))
	}
}

func (h *handler) GetLanguages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		languages, err := h.lookupService.GetLanguages(ctx)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error getting languages", "error", err)
				render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.MapToGetLanguagesResponse(languages))
	}
}

func (h *handler) GetReligions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		religions, err := h.lookupService.GetReligions(ctx)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error getting religions", "error", err)
				render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.MapToGetReligionsResponse(religions))
	}
}

func (h *handler) GetPoliticalBeliefs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		beliefs, err := h.lookupService.GetPoliticalBeliefs(ctx)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error getting political beliefs", "error", err)
				render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.MapToGetPoliticalBeliefsResponse(beliefs))
	}
}
