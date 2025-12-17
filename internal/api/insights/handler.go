package insights

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
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
	GetRetentionStats() http.HandlerFunc
	GetRetentionCohorts() http.HandlerFunc
	GetUserRetentionProfile() http.HandlerFunc
	GetGlobalRetentionStats() http.HandlerFunc
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

func (h *handler) GetRetentionStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Parse query parameters
		fromStr := r.URL.Query().Get("from")
		toStr := r.URL.Query().Get("to")

		if fromStr == "" || toStr == "" {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("from and to query parameters are required (format: YYYY-MM-DD)"))
			return
		}

		from, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid from date format, expected YYYY-MM-DD"))
			return
		}

		to, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid to date format, expected YYYY-MM-DD"))
			return
		}

		result, err := h.insightsSvc.GetRetentionStats(ctx, from, to)
		if err != nil {
			h.logger.Sugar().Warnw("get retention stats failed", "error", err, "from", from, "to", to)
			render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

			return
		}

		render.Json(w, http.StatusOK, dto.MapRetentionStats(result))
	}
}

func (h *handler) GetRetentionCohorts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Parse query parameters
		signupDateStr := r.URL.Query().Get("signup_date")
		daysAfterStr := r.URL.Query().Get("days_after")

		if signupDateStr == "" {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("signup_date query parameter is required (format: YYYY-MM-DD)"))
			return
		}

		signupDate, err := time.Parse("2006-01-02", signupDateStr)
		if err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid signup_date format, expected YYYY-MM-DD"))
			return
		}

		daysAfter := 7 // default

		if daysAfterStr != "" {
			// Parse as integer
			if d, err := strconv.Atoi(daysAfterStr); err == nil {
				daysAfter = d
			}
		}

		result, err := h.insightsSvc.GetRetentionCohorts(ctx, signupDate, daysAfter)
		if err != nil {
			h.logger.Sugar().Warnw("get retention cohorts failed", "error", err, "signupDate", signupDate, "daysAfter", daysAfter)
			render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

			return
		}

		render.Json(w, http.StatusOK, dto.MapRetentionCohort(result))
	}
}

func (h *handler) GetUserRetentionProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := chi.URLParam(r, "userID")
		if userID == "" {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("userID is required"))
			return
		}

		result, err := h.insightsSvc.GetUserRetentionProfile(ctx, userID)
		if err != nil {
			h.logger.Sugar().Warnw("get user retention profile failed", "error", err, "userID", userID)
			render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

			return
		}

		render.Json(w, http.StatusOK, dto.MapUserRetentionProfile(result))
	}
}

func (h *handler) GetGlobalRetentionStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		dateStr := r.URL.Query().Get("date")
		if dateStr == "" {
			// Default to today
			dateStr = time.Now().UTC().Format("2006-01-02")
		}

		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid date format, expected YYYY-MM-DD"))
			return
		}

		result, err := h.insightsSvc.GetGlobalRetentionStats(ctx, date)
		if err != nil {
			h.logger.Sugar().Warnw("get global retention stats failed", "error", err, "date", date)
			render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

			return
		}

		render.Json(w, http.StatusOK, dto.MapRetentionStats(result))
	}
}

func startOfWeek(t time.Time) time.Time {
	// Use Monday=0
	offset := (int(t.Weekday()) + 6) % 7
	day := t.AddDate(0, 0, -offset)

	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
}
