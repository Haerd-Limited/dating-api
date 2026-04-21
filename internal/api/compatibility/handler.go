package compatibility

import (
	"net/http"
	"strconv"

	"github.com/friendsofgo/errors"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/compatibility/dto"
	"github.com/Haerd-Limited/dating-api/internal/api/compatibility/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/compatibility"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	commonMessages "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	GetQuestions() http.HandlerFunc
	GetOverview() http.HandlerFunc
	SaveAnswer() http.HandlerFunc
}

type handler struct {
	logger               *zap.Logger
	compatibilityService compatibility.Service
}

func NewCompatibilityHandler(
	logger *zap.Logger,
	compatibilityService compatibility.Service,
) Handler {
	return &handler{
		logger:               logger,
		compatibilityService: compatibilityService,
	}
}

func (h *handler) GetOverview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		result, err := h.compatibilityService.GetOverview(ctx, userID)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "GetOverview", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapDomainToGetOverviewResponse(result))
	}
}

func (h *handler) SaveAnswer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.SaveAnswerRequest
		// Validates and decodes request
		err := request.DecodeAndValidate(r.Body, &req)
		if err != nil {
			h.logger.Sugar().Warnw("failed to decode and validate save answer request body", "error", err.Error())
			render.Json(
				w,
				http.StatusBadRequest,
				commonMappers.ToSimpleErrorResponse("invalid json"),
			)

			return
		}

		err = h.compatibilityService.SaveAnswer(ctx, mapper.MapSaveAnswerRequestToDomain(req, userID))
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "SaveAnswer", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Answer saved successfully"))
	}
}

func (h *handler) GetQuestions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		limit := request.ParseQueryInt(r, "limit", 5)
		offset := request.ParseQueryInt(r, "offset", 0)
		category := r.URL.Query().Get("category")
		viewAll := r.URL.Query().Get("view") == "all"

		var questionID *int64

		if qIDStr := r.URL.Query().Get("question_id"); qIDStr != "" {
			qID, err := strconv.ParseInt(qIDStr, 10, 64)
			if err != nil || qID < 1 {
				h.logger.Sugar().Warnw("invalid question_id query param", "question_id", qIDStr)
				render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid question_id"))

				return
			}

			questionID = &qID
		}

		var userIDPtr *string
		if userID != "" {
			userIDPtr = &userID
		}

		result, err := h.compatibilityService.GetQuestionsAndAnswers(ctx, category, offset, limit, userIDPtr, viewAll, questionID)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "GetQuestions", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapDomainToQuestionAndAnswerResponse(result))
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, compatibility.ErrAcceptableAnswerIDsRequired):
		return http.StatusBadRequest, "Acceptable answer ids required when importance is 'very'"
	case errors.Is(err, compatibility.ErrInvalidAnswerID):
		return http.StatusBadRequest, "Invalid answer id"
	case errors.Is(err, compatibility.ErrInvalidImportance):
		return http.StatusBadRequest, "Invalid importance"
	case errors.Is(err, compatibility.ErrSequentialAnsweringRequired):
		return http.StatusBadRequest, "Questions must be answered sequentially"
	case errors.Is(err, compatibility.ErrQuestionNotFound):
		return http.StatusNotFound, "Question not found"
	default:
		return http.StatusInternalServerError, commonMessages.InternalServerErrorMsg
	}
}
