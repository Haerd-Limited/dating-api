package auditlog

import (
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/auditlog/dto"
	internalauditlog "github.com/Haerd-Limited/dating-api/internal/auditlog"
	"github.com/Haerd-Limited/dating-api/internal/auditlog/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
)

type Handler interface {
	ListEvents() http.HandlerFunc
}

type handler struct {
	logger *zap.Logger
	svc    internalauditlog.Service
}

func NewHandler(logger *zap.Logger, svc internalauditlog.Service) Handler {
	return &handler{logger: logger, svc: svc}
}

func (h *handler) ListEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		limit := 50
		offset := 0

		if v := r.URL.Query().Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				limit = n
			}
		}

		if v := r.URL.Query().Get("offset"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n >= 0 {
				offset = n
			}
		}

		filter := domain.ListFilter{Limit: limit, Offset: offset}
		if actor := strings.TrimSpace(r.URL.Query().Get("actor")); actor != "" {
			filter.ActorName = &actor
		}

		if action := strings.TrimSpace(r.URL.Query().Get("action")); action != "" {
			filter.Action = &action
		}

		events, err := h.svc.ListEvents(ctx, filter)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "ListAdminEvents", err, nil)
			return
		}

		resp := dto.ListEventsResponse{Limit: limit, Offset: offset, Events: make([]dto.EventResponse, 0, len(events))}

		for _, e := range events {
			row := dto.EventResponse{
				Type:         e.Label,
				ResourceType: domain.ResourceTypeFromPath(e.Path),
				ResourceID:   "",
				ActorName:    "",
				Status:       strconv.Itoa(e.StatusCode),
				OccurredAt:   e.OccurredAt.UTC().Format(timeRFC3339),
			}

			if e.ActorName != nil {
				row.ActorName = *e.ActorName
			}

			if e.TargetID != nil {
				row.ResourceID = *e.TargetID
			}

			resp.Events = append(resp.Events, row)
		}

		render.Json(w, http.StatusOK, resp)
	}
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"
