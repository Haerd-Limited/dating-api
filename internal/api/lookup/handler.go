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
	GetEthnicities() http.HandlerFunc
	GetGenders() http.HandlerFunc
	GetDatingIntentions() http.HandlerFunc
	GetHabits() http.HandlerFunc
	GetEducationLevels() http.HandlerFunc
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

func (h *handler) GetEthnicities() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ethnicities, err := h.lookupService.GetEthnicities(ctx)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error getting ethnicities", "error", err)
				render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.MapToGetEthnicitiesResponse(ethnicities))
	}
}

func (h *handler) GetGenders() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		genders, err := h.lookupService.GetGenders(ctx)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error getting genders", "error", err)
				render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.MapToGetGendersResponse(genders))
	}
}

func (h *handler) GetDatingIntentions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		intentions, err := h.lookupService.GetDatingIntentions(ctx)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error getting dating intentions", "error", err)
				render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.MapToGetDatingIntentionsResponse(intentions))
	}
}

func (h *handler) GetHabits() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		habits, err := h.lookupService.GetHabits(ctx)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error getting habits", "error", err)
				render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.MapToGetHabitsResponse(habits))
	}
}

func (h *handler) GetEducationLevels() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		levels, err := h.lookupService.GetEducationLevels(ctx)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error getting education levels", "error", err)
				render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.MapToGetEducationLevelsResponse(levels))
	}
}
