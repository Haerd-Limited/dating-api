package lookup

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/lookup/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/lookup"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
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
	GetFamilyPlans() http.HandlerFunc
	GetFamilyStatus() http.HandlerFunc
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

func (h *handler) GetFamilyStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		familyPlans, err := h.lookupService.GetFamilyStatus(ctx)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetFamilyStatus", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetFamilyStatusResponse(familyPlans))
	}
}

func (h *handler) GetFamilyPlans() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		familyPlans, err := h.lookupService.GetFamilyPlans(ctx)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetFamilyPlans", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetFamilyPlansResponse(familyPlans))
	}
}

func (h *handler) GetPrompts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		prompts, err := h.lookupService.GetPrompts(ctx)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetPrompts", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetPromptsResponse(prompts))
	}
}

func (h *handler) GetLanguages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		languages, err := h.lookupService.GetLanguages(ctx)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetLanguages", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetLanguagesResponse(languages))
	}
}

func (h *handler) GetReligions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		religions, err := h.lookupService.GetReligions(ctx)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetReligions", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetReligionsResponse(religions))
	}
}

func (h *handler) GetPoliticalBeliefs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		beliefs, err := h.lookupService.GetPoliticalBeliefs(ctx)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetPoliticalBeliefs", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetPoliticalBeliefsResponse(beliefs))
	}
}

func (h *handler) GetEthnicities() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ethnicities, err := h.lookupService.GetEthnicities(ctx)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetEthnicities", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetEthnicitiesResponse(ethnicities))
	}
}

func (h *handler) GetGenders() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		genders, err := h.lookupService.GetGenders(ctx)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetGenders", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetGendersResponse(genders))
	}
}

func (h *handler) GetDatingIntentions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		intentions, err := h.lookupService.GetDatingIntentions(ctx)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetDatingIntentions", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetDatingIntentionsResponse(intentions))
	}
}

func (h *handler) GetHabits() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		habits, err := h.lookupService.GetHabits(ctx)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetHabits", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetHabitsResponse(habits))
	}
}

func (h *handler) GetEducationLevels() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		levels, err := h.lookupService.GetEducationLevels(ctx)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetEducationLevels", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapToGetEducationLevelsResponse(levels))
	}
}

func (h *handler) handleServiceErrorResponse(w http.ResponseWriter, r *http.Request, handlerName string, err error) {
	if render.ErrorCausedByTimeoutOrClientCancellation(w, r, h.logger, err) {
		return
	}

	h.logger.Sugar().Errorw(fmt.Sprintf("%s failure", handlerName), "error", err)
	render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))
}
