package feedback

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	dto "github.com/Haerd-Limited/dating-api/internal/api/feedback/dto"
	dtoMapper "github.com/Haerd-Limited/dating-api/internal/api/feedback/dto/mapper"
	internalfeedback "github.com/Haerd-Limited/dating-api/internal/feedback"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	CreateFeedback() http.HandlerFunc
}

type handler struct {
	logger          *zap.Logger
	feedbackService internalfeedback.Service
}

func NewFeedbackHandler(logger *zap.Logger, feedbackService internalfeedback.Service) Handler {
	return &handler{
		logger:          logger,
		feedbackService: feedbackService,
	}
}

func (h *handler) CreateFeedback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.CreateFeedbackRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Warnf("failed to decode and validate feedback request body: %s", err.Error())
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))

			return
		}

		domainReq := dtoMapper.CreateFeedbackRequestToDomain(req, userID)

		err := h.feedbackService.CreateFeedback(ctx, domainReq)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "CreateFeedback", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusCreated, dto.CreateFeedbackResponse{
			Message: "Feedback submitted successfully",
		})
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, internalfeedback.ErrInvalidFeedbackType):
		return http.StatusBadRequest, "Invalid feedback type, must be 'positive' or 'negative'"
	case errors.Is(err, internalfeedback.ErrTitleRequiredForNegative):
		return http.StatusBadRequest, "Title is required for negative feedback"
	case errors.Is(err, internalfeedback.ErrAttachmentsOnlyForNegative):
		return http.StatusBadRequest, "Attachments are only allowed for negative feedback"
	case errors.Is(err, internalfeedback.ErrTextRequired):
		return http.StatusBadRequest, "Text is required"
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
