package insights

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/insights/dto"
	"github.com/Haerd-Limited/dating-api/internal/insights"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
)

type Handler interface {
	GetPublicWeekly() http.HandlerFunc
	GetMeWeekly() http.HandlerFunc
	GetMyWrapped() http.HandlerFunc
}

type handler struct {
	logger      *zap.Logger
	insightsSvc insights.Service
}

func NewHandler(logger *zap.Logger, svc insights.Service) Handler {
	return &handler{
		logger:      logger,
		insightsSvc: svc,
	}
}

func (h *handler) GetPublicWeekly() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		weekStart := startOfWeek(time.Now().UTC())

		result, err := h.insightsSvc.GetPublicWeekly(r.Context(), weekStart)
		if err != nil {
			h.logger.Sugar().Warnw("get public weekly insights failed", "error", err)
			render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

			return
		}

		render.Json(w, http.StatusOK, dto.MapGlobalWeekly(result))
	}
}

func (h *handler) GetMeWeekly() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		weekStart := startOfWeek(time.Now().UTC())

		result, err := h.insightsSvc.GetUserWeekly(ctx, userID, weekStart)
		if err != nil {
			h.logger.Sugar().Warnw("get user weekly insights failed", "error", err, "userID", userID)
			render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

			return
		}

		render.Json(w, http.StatusOK, dto.MapUserWeekly(result))
	}
}

func (h *handler) GetMyWrapped() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		year := time.Now().UTC().Year()

		if v := r.URL.Query().Get("year"); v != "" {
			if parsed, err := time.Parse("2006", v); err == nil {
				year = parsed.Year()
			}
		}

		result, err := h.insightsSvc.GetWrapped(ctx, userID, year)
		if err != nil {
			h.logger.Sugar().Warnw("get wrapped failed", "error", err, "userID", userID, "year", year)
			render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

			return
		}

		render.Json(w, http.StatusOK, dto.MapWrapped(result))
	}
}

func startOfWeek(t time.Time) time.Time {
	// Use Monday=0
	offset := (int(t.Weekday()) + 6) % 7
	day := t.AddDate(0, 0, -offset)

	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
}
