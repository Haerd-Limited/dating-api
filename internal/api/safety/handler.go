package safety

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	dto "github.com/Haerd-Limited/dating-api/internal/api/safety/dto"
	dtoMapper "github.com/Haerd-Limited/dating-api/internal/api/safety/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/safety"
	safetydomain "github.com/Haerd-Limited/dating-api/internal/safety/domain"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	Block() http.HandlerFunc
	Report() http.HandlerFunc
	AdminListReports() http.HandlerFunc
	AdminGetReport() http.HandlerFunc
	AdminResolveReport() http.HandlerFunc
}

type handler struct {
	logger        *zap.Logger
	safetyService safety.Service
}

func NewHandler(logger *zap.Logger, safetyService safety.Service) Handler {
	return &handler{
		logger:        logger,
		safetyService: safetyService,
	}
}

func (h *handler) Block() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.BlockRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Warnf("failed to decode and validate block request body : %s", err.Error())
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))

			return
		}

		domainReq := dtoMapper.BlockRequestToDomain(req, userID)

		err := h.safetyService.BlockUser(ctx, domainReq)
		if err != nil {
			status, msg := mapBlockError(err)
			render.Json(w, status, map[string]string{"error": msg})

			return
		}

		render.Json(w, http.StatusOK, dto.BlockResponse{
			TargetUserID: req.TargetUserID,
			Status:       "blocked",
		})
	}
}

func (h *handler) Report() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.ReportRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Warnf("failed to decode and validate report request body : %s", err.Error())

			if strings.Contains(err.Error(), "ReportRequest.SubjectType") && strings.Contains(err.Error(), "oneof") {
				render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("subject_type must be one of: user, message, profile"))
				return
			}

			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))

			return
		}

		domainReq := dtoMapper.ReportRequestToDomain(req, userID)

		reportID, err := h.safetyService.CreateReport(ctx, domainReq)
		if err != nil {
			status, msg := mapReportError(err)
			render.Json(w, status, commonMappers.ToSimpleErrorResponse(msg))

			return
		}

		render.Json(w, http.StatusCreated, map[string]string{
			"id":     reportID,
			"status": string(safetydomain.ReportStatusPending),
		})
	}
}

func (h *handler) AdminListReports() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var filter safetydomain.ReportListFilter

		if statuses := r.URL.Query().Get("status"); statuses != "" {
			for _, st := range strings.Split(statuses, ",") {
				st = strings.TrimSpace(st)
				if st == "" {
					continue
				}

				status, parseErr := parseReportStatus(st)
				if parseErr != nil {
					render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse(parseErr.Error()))
					return
				}

				filter.Status = append(filter.Status, status)
			}
		}

		if categories := r.URL.Query().Get("category"); categories != "" {
			filter.Category = strings.Split(categories, ",")
		}

		if reporter := strings.TrimSpace(r.URL.Query().Get("reporter_id")); reporter != "" {
			filter.Reporter = &reporter
		}

		if reported := strings.TrimSpace(r.URL.Query().Get("reported_id")); reported != "" {
			filter.Reported = &reported
		}

		filter.Limit = request.ParseQueryInt(r, "limit", 50)
		filter.Offset = request.ParseQueryInt(r, "offset", 0)

		reports, err := h.safetyService.ListReports(ctx, filter)
		if err != nil {
			h.logger.Sugar().Errorw("list reports", "error", err)
			render.Json(w, http.StatusInternalServerError, map[string]string{"error": messages.InternalServerErrorMsg})

			return
		}

		render.Json(w, http.StatusOK, dtoMapper.MapReportsDomainToDTO(reports))
	}
}

func (h *handler) AdminGetReport() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		reportID := chi.URLParam(r, "reportID")
		if strings.TrimSpace(reportID) == "" {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("reportID is required"))
			return
		}

		report, err := h.safetyService.GetReport(ctx, reportID)
		if err != nil {
			h.logger.Sugar().Errorw("get report", "reportID", reportID, "error", err)
			render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

			return
		}

		if report == nil {
			render.Json(w, http.StatusNotFound, commonMappers.ToSimpleErrorResponse("report not found"))
			return
		}

		render.Json(w, http.StatusOK, dtoMapper.MapReportDomainToDTO(*report))
	}
}

func (h *handler) AdminResolveReport() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		reportID := chi.URLParam(r, "reportID")
		if strings.TrimSpace(reportID) == "" {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("reportID is required"))
			return
		}

		adminID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.ResolveReportRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			render.Json(w, http.StatusBadRequest, map[string]string{
				"error": "invalid request payload",
			})

			return
		}

		domainReq, err := dtoMapper.ResolveReportRequestToDomain(req, reportID, adminID)
		if err != nil {
			render.Json(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		err = h.safetyService.ResolveReport(ctx, domainReq)
		if err != nil {
			if errors.Is(err, safety.ErrReportNotFound) {
				render.Json(w, http.StatusNotFound, map[string]string{"error": "report not found"})
				return
			}

			h.logger.Sugar().Errorw("resolve report", "reportID", reportID, "error", err)
			render.Json(w, http.StatusInternalServerError, map[string]string{"error": messages.InternalServerErrorMsg})

			return
		}

		render.Json(w, http.StatusOK, map[string]string{
			"status": string(req.NewStatus),
		})
	}
}

func mapBlockError(err error) (int, string) {
	switch {
	case errors.Is(err, safety.ErrSelfBlock):
		return http.StatusBadRequest, "you cannot block yourself"
	case errors.Is(err, safety.ErrInvalidBlockRequest):
		return http.StatusBadRequest, "blocker_id and blocked_id are required"
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}

func mapReportError(err error) (int, string) {
	switch {
	case errors.Is(err, safety.ErrInvalidReportRequest):
		return http.StatusBadRequest, "reporter_id and reported_user_id are required"
	case errors.Is(err, safety.ErrSelfReport):
		return http.StatusBadRequest, "you cannot report yourself"
	case errors.Is(err, safety.ErrReportNotFound):
		return http.StatusNotFound, "report not found"
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}

func parseReportStatus(raw string) (safetydomain.ReportStatus, error) {
	status := safetydomain.ReportStatus(strings.ToLower(strings.TrimSpace(raw)))
	switch status {
	case safetydomain.ReportStatusPending,
		safetydomain.ReportStatusInReview,
		safetydomain.ReportStatusResolved,
		safetydomain.ReportStatusEscalated,
		safetydomain.ReportStatusDismissed:
		return status, nil
	default:
		return "", fmt.Errorf("invalid status: %s", raw)
	}
}
