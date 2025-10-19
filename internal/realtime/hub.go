package realtime

import (
	"sync"
)

type Broadcaster interface {
	BroadcastToConversation(convoID string, payload []byte)
	BroadcastToUser(userID string, payload []byte)
}

type Hub struct {
	mu             sync.RWMutex
	byUser         map[string]map[*Conn]struct{} // userID -> conns
	byConversation map[string]map[*Conn]struct{} // convoID -> conns
}

func NewHub() *Hub {
	return &Hub{
		byUser:         make(map[string]map[*Conn]struct{}),
		byConversation: make(map[string]map[*Conn]struct{}),
	}
}

func (h *Hub) BroadcastToUser(userID string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.byUser[userID] {
		c.Enqueue(payload)
	}
}

func (h *Hub) Add(c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	m := h.byUser[c.UserID]
	if m == nil {
		m = make(map[*Conn]struct{})
		h.byUser[c.UserID] = m
	}

	m[c] = struct{}{}
}

func (h *Hub) Remove(c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if m, ok := h.byUser[c.UserID]; ok {
		delete(m, c)

		if len(m) == 0 {
			delete(h.byUser, c.UserID)
		}
	}
	// also remove from byConversation sets where present (call from Conn.Close)
}

func (h *Hub) JoinConversation(c *Conn, convoID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	m := h.byConversation[convoID]
	if m == nil {
		m = make(map[*Conn]struct{})
		h.byConversation[convoID] = m
	}

	m[c] = struct{}{}
}

func (h *Hub) LeaveConversation(c *Conn, convoID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if m, ok := h.byConversation[convoID]; ok {
		delete(m, c)

		if len(m) == 0 {
			delete(h.byConversation, convoID)
		}
	}
}

func (h *Hub) BroadcastToConversation(convoID string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.byConversation[convoID] {
		c.Enqueue(payload)
	}
}
