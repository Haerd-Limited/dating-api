package matching

import (
	"fmt"
	"net/http"

	"github.com/friendsofgo/errors"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/matching/dto"
	"github.com/Haerd-Limited/dating-api/internal/api/matching/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/matching"
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
	logger          *zap.Logger
	matchingService matching.Service
}

func NewMatchingHandler(
	logger *zap.Logger,
	matchingService matching.Service,
) Handler {
	return &handler{
		logger:          logger,
		matchingService: matchingService,
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

		result, err := h.matchingService.GetOverview(ctx, userID)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetOverview", err)
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

		err = h.matchingService.SaveAnswer(ctx, mapper.MapSaveAnswerRequestToDomain(req, userID))
		if err != nil {
			h.handleServiceErrorResponse(w, r, "SaveAnswer", err)
			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Answer saved successfully"))
	}
}

func (h *handler) GetQuestions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		_, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		limit := request.ParseQueryInt(r, "limit", 5)
		offset := request.ParseQueryInt(r, "offset", 0)
		category := r.URL.Query().Get("category")

		result, err := h.matchingService.GetQuestionsAndAnswers(ctx, category, offset, limit)
		if err != nil {
			h.handleServiceErrorResponse(w, r, "GetQuestions", err)
			return
		}

		render.Json(w, http.StatusOK, mapper.MapDomainToQuestionAndAnswerResponse(result))
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
	case errors.Is(err, matching.ErrAcceptableAnswerIDsRequired):
		return http.StatusBadRequest, "Acceptable answer ids required when importance is 'very'"
	case errors.Is(err, matching.ErrInvalidAnswerID):
		return http.StatusBadRequest, "Invalid answer id"
	case errors.Is(err, matching.ErrInvalidImportance):
		return http.StatusBadRequest, "Invalid importance"
	default:
		return http.StatusInternalServerError, commonMessages.InternalServerErrorMsg
	}
}
