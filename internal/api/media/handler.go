package media

import (
	"context"
	"errors"
	"github.com/Haerd-Limited/dating-api/internal/api/media/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/media"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"go.uber.org/zap"
	"net/http"
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
			authHeader := r.Header.Get("Authorization")
			h.logger.Sugar().Errorw("missing user ID", "authHeader", authHeader)

			render.Json(
				w,
				http.StatusUnauthorized,
				commonMappers.ToSimpleErrorResponse(
					messages.AuthenticationRequiredMsg,
				))

			return
		}

		url, err := h.mediaService.GeneratePhotoUploadUrl(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error generating photo upload url", "error", err)
				//statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.MapUploadURLToResponse(url))
	}
}

func (h *handler) GenerateVoiceNoteUploadUrl() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			authHeader := r.Header.Get("Authorization")
			h.logger.Sugar().Errorw("missing user ID", "authHeader", authHeader)

			render.Json(
				w,
				http.StatusUnauthorized,
				commonMappers.ToSimpleErrorResponse(
					messages.AuthenticationRequiredMsg,
				))

			return
		}

		url, err := h.mediaService.GenerateVoiceNoteUploadUrl(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error generating voicenote upload url", "error", err)
				//statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, http.StatusInternalServerError, commonMappers.ToSimpleErrorResponse(messages.InternalServerErrorMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.MapUploadURLToResponse(url))
	}
}
