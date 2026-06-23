package adminrealtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/adminrealtime"
	"github.com/Haerd-Limited/dating-api/internal/api/adminrealtime/dto"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/request"
)

type Handler interface {
	Stream() http.HandlerFunc
	SetPresence() http.HandlerFunc
	ClearPresence() http.HandlerFunc
}

type handler struct {
	logger   *zap.Logger
	hub      *adminrealtime.Hub
	presence *adminrealtime.PresenceStore
}

func NewHandler(logger *zap.Logger, hub *adminrealtime.Hub, presence *adminrealtime.PresenceStore) Handler {
	return &handler{logger: logger, hub: hub, presence: presence}
}

func (h *handler) Stream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, displayName, ok := commoncontext.AdminActorFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		rc := http.NewResponseController(w)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		sub := h.hub.Subscribe(sessionID, displayName)
		defer func() {
			h.hub.Unsubscribe(sub)

			for _, entry := range h.presence.ReleaseSession(sessionID) {
				h.hub.BroadcastEvent(adminrealtime.Event{
					Type:         adminrealtime.EventPresenceReleased,
					ResourceType: entry.ResourceType,
					ResourceID:   entry.ResourceID,
					ActorName:    entry.DisplayName,
				})
			}
		}()

		for _, entry := range h.presence.Snapshot() {
			h.writeSSE(w, rc, adminrealtime.Event{
				Type:         adminrealtime.EventPresenceUpdated,
				ResourceType: entry.ResourceType,
				ResourceID:   entry.ResourceID,
				ActorName:    entry.DisplayName,
				OccurredAt:   entry.Since,
			})
		}

		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()

		ctx := r.Context()

		for {
			select {
			case <-ctx.Done():
				return
			case payload, open := <-sub.Ch:
				if !open {
					return
				}

				if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
					return
				}

				if err := rc.Flush(); err != nil {
					return
				}
			case <-ticker.C:
				if _, err := fmt.Fprintf(w, ": keep-alive\n\n"); err != nil {
					return
				}

				if err := rc.Flush(); err != nil {
					return
				}
			}
		}
	}
}

func (h *handler) SetPresence() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, displayName, ok := commoncontext.AdminActorFromContext(r.Context())
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.PresenceRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			render.Json(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}

		entry := h.presence.Set(sessionID, displayName, req.ResourceType, req.ResourceID)
		h.hub.BroadcastEvent(adminrealtime.Event{
			Type:         adminrealtime.EventPresenceUpdated,
			ResourceType: entry.ResourceType,
			ResourceID:   entry.ResourceID,
			ActorName:    entry.DisplayName,
			OccurredAt:   entry.Since,
		})

		render.Json(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func (h *handler) ClearPresence() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, displayName, ok := commoncontext.AdminActorFromContext(r.Context())
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		var req dto.PresenceRequest
		if err := request.DecodeAndValidate(r.Body, &req); err != nil {
			render.Json(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}

		entry, cleared := h.presence.Clear(sessionID, req.ResourceType, req.ResourceID)
		if cleared {
			h.hub.BroadcastEvent(adminrealtime.Event{
				Type:         adminrealtime.EventPresenceReleased,
				ResourceType: entry.ResourceType,
				ResourceID:   entry.ResourceID,
				ActorName:    displayName,
			})
		}

		render.Json(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func (h *handler) writeSSE(w http.ResponseWriter, rc *http.ResponseController, evt adminrealtime.Event) {
	if evt.OccurredAt.IsZero() {
		evt.OccurredAt = time.Now().UTC()
	}

	b, err := json.Marshal(evt)
	if err != nil {
		return
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", b); err != nil {
		return
	}

	_ = rc.Flush()
}
