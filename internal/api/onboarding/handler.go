package onboarding

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/onboarding"
)

type Handler interface {
	Patch() http.HandlerFunc
	PatchVisibility() http.HandlerFunc
	Complete() http.HandlerFunc
	State() http.HandlerFunc
}

type handler struct {
	logger            *zap.Logger
	onboardingService onboarding.Service
}

func NewOnboardingHandler(
	logger *zap.Logger,
	onboardingService onboarding.Service,
) Handler {
	return &handler{
		logger:            logger,
		onboardingService: onboardingService,
	}
}

func (h *handler) Patch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
	}
}

func (h *handler) PatchVisibility() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (h *handler) Complete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (h *handler) State() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
