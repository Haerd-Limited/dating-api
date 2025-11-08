package verification

import (
	"errors"
	"fmt"
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
			h.handleServiceErrorResponse(w, r, "Start", err)
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
			h.handleServiceErrorResponse(w, r, "Complete", err)
			return
		}

		render.Json(w, http.StatusOK, dto.MapToCompleteResponse(result))
	}
}

func (h *handler) handleServiceErrorResponse(w http.ResponseWriter, r *http.Request, handlerName string, err error) {
	if render.ErrorCausedByTimeoutOrClientCancellation(w, r, h.logger, err) {
		return
	}

	statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
	if statusCode == http.StatusInternalServerError {
		h.logger.Sugar().Errorw(fmt.Sprintf("%s failure", handlerName), "error", err.Error())
	} else {
		h.logger.Sugar().Warnw(fmt.Sprintf("%s failure", handlerName), "error", err.Error())
	}

	render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, commonErrors.ErrInvalidMediaUrl):
		return http.StatusBadRequest, "Invalid media url"
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
