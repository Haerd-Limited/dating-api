package realtime

import (
	"net/http"

	"github.com/coder/websocket"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/realtime"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
)

type Handler interface {
	ServeWS() http.HandlerFunc
}

type wsHandler struct {
	Hub    *realtime.Hub // connection manager (below)
	logger *zap.Logger
	auth   realtime.ConversationAuth
}

func NewWsHandler(
	logger *zap.Logger,
	hub *realtime.Hub,
	auth realtime.ConversationAuth,
) Handler {
	return &wsHandler{
		logger: logger,
		Hub:    hub,
		auth:   auth,
	}
}

func (h *wsHandler) ServeWS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := commoncontext.UserIDFromContext(ctx)
		if !ok {
			render.UnauthorizedResponse(w, r, h.logger)
			return
		}

		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}

		conn := realtime.NewConn(userID, c, h.Hub, h.logger, h.auth)
		h.Hub.Add(conn)

		defer func() {
			conn.Close()
			h.Hub.Remove(conn)
		}()

		// read loop: handle client control messages (subscribe/unsubscribe/ping)
		conn.ReadLoop(r.Context())
	}
}
