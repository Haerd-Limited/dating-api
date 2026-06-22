package consent

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/consent/dto"
	dtoMapper "github.com/Haerd-Limited/dating-api/internal/api/consent/dto/mapper"
	internalconsent "github.com/Haerd-Limited/dating-api/internal/consent"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	commonMappers "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/mappers"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	Record() http.HandlerFunc
	List() http.HandlerFunc
	Revoke() http.HandlerFunc
}

type handler struct {
	logger         *zap.Logger
	consentService internalconsent.Service
}

func NewConsentHandler(logger *zap.Logger, consentService internalconsent.Service) Handler {
	return &handler{
		logger:         logger,
		consentService: consentService,
	}
}

func (h *handler) Record() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.RecordConsentRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			h.logger.Sugar().Warnf("failed to decode consent request: %s", err.Error())
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("invalid request payload"))

			return
		}

		ip := request.ClientIP(r)
		userAgent := r.Header.Get("User-Agent")
		domainReq := dtoMapper.RecordRequestToDomain(req, userID, &ip, &userAgent)

		err := h.consentService.Record(ctx, domainReq)
		if err != nil {
			if isUniqueViolation(err) {
				render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Consent already recorded"))
				return
			}

			render.HandleServiceErrorResponse(h.logger, w, r, "RecordConsent", err, mapErrorsToStatusCodeAndUserFriendlyMessages)

			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Consent recorded"))
	}
}

func (h *handler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		consents, err := h.consentService.ListForUser(ctx, userID)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "ListConsents", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, dtoMapper.DomainsToResponse(consents))
	}
}

func (h *handler) Revoke() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		consentType := chi.URLParam(r, "type")
		if consentType == "" {
			render.Json(w, http.StatusBadRequest, commonMappers.ToSimpleErrorResponse("consent type is required"))
			return
		}

		err := h.consentService.Revoke(ctx, userID, consentType)
		if err != nil {
			render.HandleServiceErrorResponse(h.logger, w, r, "RevokeConsent", err, mapErrorsToStatusCodeAndUserFriendlyMessages)
			return
		}

		render.Json(w, http.StatusOK, commonMappers.ToSimpleMessageResponse("Consent revoked"))
	}
}

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, internalconsent.ErrInvalidConsentType):
		return http.StatusBadRequest, "Invalid consent type"
	case errors.Is(err, internalconsent.ErrConsentVersion):
		return http.StatusBadRequest, "Invalid consent version"
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}

	return strings.Contains(err.Error(), "duplicate key")
}
