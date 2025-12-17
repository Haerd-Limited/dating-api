package verification

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/verification/dto"
	"github.com/Haerd-Limited/dating-api/internal/verification"
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

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, commonErrors.ErrInvalidMediaUrl):
		return http.StatusBadRequest, "Invalid media url"
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
