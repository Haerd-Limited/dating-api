package media

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/media/dto"
	"github.com/Haerd-Limited/dating-api/internal/api/media/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/media"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	GeneratePhotoUploadUrl() http.HandlerFunc
	GenerateVoiceNoteUploadUrl() http.HandlerFunc
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

		var req dto.GenerateVoiceNoteUploadUrlRequest

		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate generate voicenote upload url request body", "error", err)
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse("All fields are required and action field must be one of 'voicenote' or 'prompt'"),
			)

			return
		}

		url, err := h.mediaService.GenerateVoiceNoteUploadUrl(ctx, userID, req.Purpose)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GenerateVoiceNoteUploadUrl", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapUploadURLToResponse(url))
	}
}

func (h *handler) handleServiceErrorResponse(w http.ResponseWriter, r *http.Request, handlerName string, err error) {
	if render.ErrorCausedByTimeoutOrClientCancellation(w, r, h.logger, err) {
		return
	}

	h.logger.Sugar().Errorw(fmt.Sprintf("%s failure", handlerName), "error", err.Error())
	render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))
}
