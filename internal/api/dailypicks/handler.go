package dailypicks

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/dailypicks/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/dailypicks"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
)

type Handler interface {
	GetDailyPicks() http.HandlerFunc
}

type handler struct {
	logger  *zap.Logger
	service dailypicks.Service
}

func NewHandler(logger *zap.Logger, service dailypicks.Service) Handler {
	return &handler{
		logger:  logger,
		service: service,
	}
}

func (h *handler) GetDailyPicks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		result, err := h.service.GetDailyPicks(ctx, userID)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "GetDailyPicks", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, mapper.DomainToGetDailyPicksResponse(result))
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	return http.StatusInternalServerError, "Something went wrong. Please try again."
}
