package verification

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/verification/dto"
	"github.com/Haerd-Limited/dating-api/internal/verification"
	verificationdomain "github.com/Haerd-Limited/dating-api/internal/verification/domain"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	Start() http.HandlerFunc
	Complete() http.HandlerFunc
	StartVideo() http.HandlerFunc
	SubmitVideo() http.HandlerFunc
	// Admin methods
	AdminListVideoAttempts() http.HandlerFunc
	AdminGetVideoAttempt() http.HandlerFunc
	AdminGetVideoDownloadURL() http.HandlerFunc
	AdminApproveVideoAttempt() http.HandlerFunc
	AdminRejectVideoAttempt() http.HandlerFunc
}

type handler struct {
	logger              *zap.Logger
	verificationService verification.Service
}

func NewVerificationHandler(
	logger *zap.Logger,
	verificationService verification.Service,
) Handler {
	return &handler{
		logger:              logger,
		verificationService: verificationService,
	}
}

func (h *handler) Start() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		result, err := h.verificationService.StartPhotoVerification(ctx, userID)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "Start", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, dto.MapToStartResponse(result))
	}
}

func (h *handler) Complete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.CompleteRequest
		// Validates and decodes request
		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate complete request body", "error", err.Error())
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse("session_id is required"),
			)

			return
		}

		result, err := h.verificationService.CompletePhotoVerification(ctx, userID, req.SessionID)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "Complete", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, dto.MapToCompleteResponse(result))
	}
}

func (h *handler) StartVideo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		result, err := h.verificationService.StartVideoVerification(ctx, userID)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "StartVideo", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, dto.MapToStartVideoResponse(result))
	}
}

func (h *handler) SubmitVideo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.SubmitVideoRequest

		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate submit video request body", "error", err.Error())
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse("video_key is required"),
			)

			return
		}

		result, err := h.verificationService.SubmitVideoVerification(ctx, userID, req.VideoKey)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "SubmitVideo", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, dto.MapToSubmitVideoResponse(result))
	}
}

func (h *handler) AdminListVideoAttempts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var filter verificationdomain.VideoAttemptFilter

		if statuses := r.URL.Query().Get("status"); statuses != "" {
			for _, st := range strings.Split(statuses, ",") {
				st = strings.TrimSpace(st)
				if st != "" {
					filter.Status = append(filter.Status, st)
				}
			}
		}

		if userID := strings.TrimSpace(r.URL.Query().Get("user_id")); userID != "" {
			filter.UserID = &userID
		}

		filter.Limit = request.ParseQueryInt(r, "limit", 50)
		filter.Offset = request.ParseQueryInt(r, "offset", 0)

		attempts, err := h.verificationService.ListVideoAttempts(ctx, filter)
		if err != nil {
			h.logger.Sugar().Errorw("list video attempts", "error", err)
			render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

			return
		}

		render.Json(w, http.StatusOK, dto.DomainToVideoAttemptListResponse(attempts))
	}
}

func (h *handler) AdminGetVideoAttempt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		attemptID := chi.URLParam(r, "attemptID")
		if strings.TrimSpace(attemptID) == "" {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("attemptID is required"))
			return
		}

		attempt, err := h.verificationService.GetVideoAttempt(ctx, attemptID)
		if err != nil {
			if errors.Is(err, verification.ErrVideoAttemptNotFound) {
				render.Json(w, http.StatusNotFound, commonMappers.ToSimpleErrorResponse("video attempt not found"))
				return
			}

			h.logger.Sugar().Errorw("get video attempt", "attemptID", attemptID, "error", err)
			render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

			return
		}

		render.Json(w, http.StatusOK, dto.DomainToVideoAttemptResponse(*attempt))
	}
}

func (h *handler) AdminGetVideoDownloadURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		attemptID := chi.URLParam(r, "attemptID")
		if strings.TrimSpace(attemptID) == "" {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("attemptID is required"))
			return
		}

		// Get the attempt to retrieve video S3 key
		attempt, err := h.verificationService.GetVideoAttempt(ctx, attemptID)
		if err != nil {
			if errors.Is(err, verification.ErrVideoAttemptNotFound) {
				render.Json(w, http.StatusNotFound, commonMappers.ToSimpleErrorResponse("video attempt not found"))
				return
			}

			h.logger.Sugar().Errorw("get video attempt", "attemptID", attemptID, "error", err)
			render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

			return
		}

		if attempt.VideoS3Key == "" {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("video S3 key not found for this attempt"))
			return
		}

		videoURL, err := h.verificationService.GetVideoDownloadURL(ctx, attempt.VideoS3Key)
		if err != nil {
			h.logger.Sugar().Errorw("generate video download URL", "attemptID", attemptID, "error", err)
			render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

			return
		}

		render.Json(w, http.StatusOK, dto.VideoDownloadURLResponse{
			VideoURL:  videoURL,
			ExpiresIn: 3600, // 1 hour in seconds
		})
	}
}

func (h *handler) AdminApproveVideoAttempt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		attemptID := chi.URLParam(r, "attemptID")
		if strings.TrimSpace(attemptID) == "" {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("attemptID is required"))
			return
		}

		var req dto.ApproveVideoRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Warnf("failed to decode and validate approve request body: %s", err.Error())
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))

			return
		}

		err := h.verificationService.ApproveVideoAttempt(ctx, attemptID, req.Notes)
		if err != nil {
			if errors.Is(err, verification.ErrVideoAttemptNotFound) {
				render.Json(w, http.StatusNotFound, commonMappers.ToSimpleErrorResponse("video attempt not found"))
				return
			}

			if errors.Is(err, verification.ErrInvalidVideoAttemptStatus) {
				render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse(err.Error()))
				return
			}

			h.logger.Sugar().Errorw("approve video attempt", "attemptID", attemptID, "error", err)
			render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

			return
		}

		render.Json(w, http.StatusOK, dto.ApproveVideoResponse{
			Status:  "passed",
			Message: "Video verification approved",
		})
	}
}

func (h *handler) AdminRejectVideoAttempt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		attemptID := chi.URLParam(r, "attemptID")
		if strings.TrimSpace(attemptID) == "" {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("attemptID is required"))
			return
		}

		var req dto.RejectVideoRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Warnf("failed to decode and validate reject request body: %s", err.Error())
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))

			return
		}

		err := h.verificationService.RejectVideoAttempt(ctx, attemptID, req.RejectionReason, req.Notes)
		if err != nil {
			if errors.Is(err, verification.ErrVideoAttemptNotFound) {
				render.Json(w, http.StatusNotFound, commonMappers.ToSimpleErrorResponse("video attempt not found"))
				return
			}

			if errors.Is(err, verification.ErrInvalidVideoAttemptStatus) {
				render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse(err.Error()))
				return
			}

			if errors.Is(err, verification.ErrRejectionReasonRequired) {
				render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("rejection reason is required"))
				return
			}

			h.logger.Sugar().Errorw("reject video attempt", "attemptID", attemptID, "error", err)
			render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

			return
		}

		render.Json(w, http.StatusOK, dto.RejectVideoResponse{
			Status:  "failed",
			Message: "Video verification rejected",
		})
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, commonErrors.ErrInvalidMediaUrl):
		return http.StatusBadRequest, "Invalid media url"
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
