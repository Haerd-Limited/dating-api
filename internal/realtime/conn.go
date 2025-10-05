package realtime

import (
	"context"
	"encoding/json"

	"github.com/coder/websocket"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/realtime/dto"
)

type Conn struct {
	UserID string
	ws     *websocket.Conn
	hub    *Hub
	log    *zap.Logger
	send   chan []byte
	rooms  map[string]struct{} // joined convoIDs
	auth   ConversationAuth
}

func NewConn(userID string, ws *websocket.Conn, hub *Hub, log *zap.Logger, auth ConversationAuth) *Conn {
	c := &Conn{
		UserID: userID,
		ws:     ws,
		hub:    hub,
		log:    log,
		send:   make(chan []byte, 256),
		rooms:  map[string]struct{}{},
		auth:   auth,
	}
	go c.writeLoop()

	return c
}

func (c *Conn) Enqueue(b []byte) {
	select {
	case c.send <- b:
	default:
		c.log.Warn("ws backpressure: dropping message", zap.String("user", c.UserID))
	}
}

func (c *Conn) writeLoop() {
	ctx := context.Background()
	for msg := range c.send {
		_ = c.ws.Write(ctx, websocket.MessageText, msg)
	}
}

func (c *Conn) ReadLoop(ctx context.Context) {
	for {
		_, data, err := c.ws.Read(ctx)
		if err != nil {
			return
		}
		// handle client messages: subscribe/unsubscribe/ping/typing
		var m dto.ClientMsg
		if json.Unmarshal(data, &m) == nil {
			switch m.Type {
			case "subscribe.conversation":
				ok, err2 := c.auth.IsConversationParticipant(ctx, m.ConversationID, c.UserID)
				if err2 != nil {
					c.Enqueue([]byte(`{"type":"error","code":"authz_failed"}`))
					continue
				}

				if !ok {
					c.Enqueue([]byte(`{"type":"error","code":"not_participant"}`))
					// optional: close after an authz violation
					_ = c.ws.Close(websocket.StatusPolicyViolation, "not participant")

					continue
				}

				c.hub.JoinConversation(c, m.ConversationID)
				c.rooms[m.ConversationID] = struct{}{}
			case "unsubscribe.conversation":
				c.hub.LeaveConversation(c, m.ConversationID)
				delete(c.rooms, m.ConversationID)
			case "ping":
				c.Enqueue([]byte(`{"type":"pong"}`))
			}
		}
	}
}

func (c *Conn) Close() {
	// leave any joined rooms
	for id := range c.rooms {
		c.hub.LeaveConversation(c, id)
	}
	_ = c.ws.Close(websocket.StatusNormalClosure, "")
}
