package media

import (
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/media/dto"
	"github.com/Haerd-Limited/dating-api/internal/api/media/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/media"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	GeneratePhotoUploadUrl() http.HandlerFunc
	GenerateVoiceNoteUploadUrl() http.HandlerFunc
	GenerateFeedbackAttachmentUploadUrl() http.HandlerFunc
	TranscribeInstagramReel() http.HandlerFunc
}

type handler struct {
	logger       *zap.Logger
	mediaService media.Service
}

func NewMediaHandler(
	logger *zap.Logger,
	mediaService media.Service,
) Handler {
	return &handler{
		logger:       logger,
		mediaService: mediaService,
	}
}

func (h *handler) GeneratePhotoUploadUrl() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		url, err := h.mediaService.GeneratePhotoUploadUrl(ctx, userID)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GeneratePhotoUploadUrl", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapUploadURLToResponse(url))
	}
}

func (h *handler) GenerateVoiceNoteUploadUrl() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		purpose := r.URL.Query().Get("purpose")
		if purpose == "" {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("Purpose query parameter is required"))
			return
		}

		if purpose != constants.PurposeVoiceNote && purpose != constants.PurposePrompt {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("Purpose query parameter must be one of 'voicenote' or 'prompt'"))
			return
		}

		url, err := h.mediaService.GenerateVoiceNoteUploadUrl(ctx, userID, purpose)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GenerateVoiceNoteUploadUrl", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapUploadURLToResponse(url))
	}
}

func (h *handler) GenerateFeedbackAttachmentUploadUrl() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		mediaType := r.URL.Query().Get("media_type")
		if mediaType == "" {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("media_type query parameter is required"))
			return
		}

		if mediaType != "image" && mediaType != "video" {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("media_type must be 'image' or 'video'"))
			return
		}

		url, err := h.mediaService.GenerateFeedbackAttachmentUploadUrl(ctx, userID, mediaType)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GenerateFeedbackAttachmentUploadUrl", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapUploadURLToResponse(url))
	}
}

func (h *handler) TranscribeInstagramReel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dto.TranscribeReelRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate transcribe reel request", "error", err.Error())
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload: reel_url is required and must be a valid URL"))

			return
		}

		transcript, err := h.mediaService.TranscribeInstagramReel(ctx, req.ReelURL)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "TranscribeInstagramReel", err)
			return
		}

		render.Json(w, http.StatusOK, dto.TranscribeReelResponse{
			Transcript: transcript,
		})
	}
}

func (h *handler) handleServiceErrorResponse(w http.ResponseWriter, r *http.Request, handlerName string, err error) {
	if render.ErrorCausedByTimeoutOrClientCancellation(w, r, h.logger, err) {
		return
	}

	// Handle specific Instagram errors
	if errors.Is(err, media.ErrInstagramAuthRequired) {
		h.logger.Sugar().Warnw(fmt.Sprintf("%s failure", handlerName), "error", err.Error())
		render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("Unable to access Instagram reel: authentication required. The reel may be private or require login."))

		return
	}

	if errors.Is(err, media.ErrInstagramRateLimited) {
		h.logger.Sugar().Warnw(fmt.Sprintf("%s failure", handlerName), "error", err.Error())
		render.Json(w, http.StatusTooManyRequests, commonMappers.ToSimpleErrorResponse("Instagram rate limit reached. Please try again later."))

		return
	}

	h.logger.Sugar().Errorw(fmt.Sprintf("%s failure", handlerName), "error", err.Error())
	render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))
}
