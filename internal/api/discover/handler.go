package discover

import (
	"context"
	"errors"
	"github.com/Haerd-Limited/dating-api/internal/api/discover/dto/mapper"
	"github.com/Haerd-Limited/dating-api/internal/discover"
	"github.com/Haerd-Limited/dating-api/internal/user/storage"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
	"go.uber.org/zap"
	"net/http"
)

type Handler interface {
	GetDiscover() http.HandlerFunc
}

type handler struct {
	logger          *zap.Logger
	discoverService discover.Service
}

func NewDiscoverHandler(
	logger *zap.Logger,
	discoverService discover.Service,
) Handler {
	return &handler{
		logger:          logger,
		discoverService: discoverService,
	}
}

func (h *handler) GetDiscover() http.HandlerFunc {
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

		limit := request.ParseQueryInt(r, "limit", 10)
		offset := request.ParseQueryInt(r, "offset", 0)

		result, err := h.discoverService.GetDiscoverFeed(ctx, userID, limit, offset)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				h.logger.Sugar().Infow("client canceled request", "path", r.URL.Path)
				return // no need to return a response. Client socket is closed.
			case errors.Is(err, context.DeadlineExceeded):
				render.Json(w, http.StatusGatewayTimeout, commonMappers.ToSimpleErrorResponse("request timed out"))
				return
			default:
				h.logger.Sugar().Errorw("Error getting user discover feed", "error", err)
				statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
				render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))

				return
			}
		}

		render.Json(w, http.StatusOK, mapper.FeedProfilesToDto(result))
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, storage.ErrUserDoesNotExists):
		return http.StatusNotFound, "User does not exist"
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}
