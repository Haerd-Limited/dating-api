package matching

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/matching/dto"
	"github.com/Haerd-Limited/dating-api/internal/matching"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	commonMessages "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	GetQuestions() http.HandlerFunc
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

		render.Json(w, http.StatusOK, dto.MapDomainToQuestionAndAnswerResponse(result))
	}
}

func (h *handler) handleServiceErrorResponse(w http.ResponseWriter, r *http.Request, handlerName string, err error) {
	if render.ErrorCausedByTimeoutOrClientCancellation(w, r, h.logger, err) {
		return
	}

	h.logger.Sugar().Errorw(fmt.Sprintf("%s failure", handlerName), "error", err.Error())
	statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
	render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	default:
		return http.StatusInternalServerError, commonMessages.InternalServerErrorMsg
	}
}
