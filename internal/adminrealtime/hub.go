package adminrealtime

import (
	"encoding/json"
	"sync"
	"time"
)

const (
	EventPresenceUpdated     = "presence.updated"
	EventPresenceReleased    = "presence.released"
	EventVerificationUpdated = "verification.updated"
	EventReportUpdated       = "report.updated"
)

type Event struct {
	Type         string    `json:"type"`
	ResourceType string    `json:"resource_type,omitempty"`
	ResourceID   string    `json:"resource_id,omitempty"`
	ActorName    string    `json:"actor_name,omitempty"`
	Status       string    `json:"status,omitempty"`
	OccurredAt   time.Time `json:"occurred_at"`
}

type PresenceEntry struct {
	SessionID    string
	DisplayName  string
	ResourceType string
	ResourceID   string
	Since        time.Time
}

type Subscriber struct {
	SessionID   string
	DisplayName string
	Ch          chan []byte
}

type Hub struct {
	mu          sync.RWMutex
	subscribers map[*Subscriber]struct{}
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[*Subscriber]struct{}),
	}
}

func (h *Hub) Subscribe(sessionID, displayName string) *Subscriber {
	sub := &Subscriber{
		SessionID:   sessionID,
		DisplayName: displayName,
		Ch:          make(chan []byte, 32),
	}

	h.mu.Lock()
	h.subscribers[sub] = struct{}{}
	h.mu.Unlock()

	return sub
}

func (h *Hub) Unsubscribe(sub *Subscriber) {
	if sub == nil {
		return
	}

	h.mu.Lock()
	delete(h.subscribers, sub)
	h.mu.Unlock()

	close(sub.Ch)
}

func (h *Hub) Broadcast(payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for sub := range h.subscribers {
		select {
		case sub.Ch <- payload:
		default:
		}
	}
}

func (h *Hub) BroadcastEvent(evt Event) {
	if evt.OccurredAt.IsZero() {
		evt.OccurredAt = time.Now().UTC()
	}

	b, err := json.Marshal(evt)
	if err != nil {
		return
	}

	h.Broadcast(b)
}

type PresenceStore struct {
	mu        sync.RWMutex
	byKey     map[string]PresenceEntry
	bySession map[string]string
}

func NewPresenceStore() *PresenceStore {
	return &PresenceStore{
		byKey:     make(map[string]PresenceEntry),
		bySession: make(map[string]string),
	}
}

func presenceKey(resourceType, resourceID string) string {
	return resourceType + ":" + resourceID
}

func (p *PresenceStore) Set(sessionID, displayName, resourceType, resourceID string) PresenceEntry {
	key := presenceKey(resourceType, resourceID)
	entry := PresenceEntry{
		SessionID:    sessionID,
		DisplayName:  displayName,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Since:        time.Now().UTC(),
	}

	p.mu.Lock()
	if oldKey, ok := p.bySession[sessionID]; ok && oldKey != key {
		delete(p.byKey, oldKey)
	}

	p.byKey[key] = entry
	p.bySession[sessionID] = key
	p.mu.Unlock()

	return entry
}

func (p *PresenceStore) Clear(sessionID, resourceType, resourceID string) (PresenceEntry, bool) {
	key := presenceKey(resourceType, resourceID)

	p.mu.Lock()
	defer p.mu.Unlock()

	entry, ok := p.byKey[key]
	if !ok || entry.SessionID != sessionID {
		return PresenceEntry{}, false
	}

	delete(p.byKey, key)
	delete(p.bySession, sessionID)

	return entry, true
}

func (p *PresenceStore) ReleaseSession(sessionID string) []PresenceEntry {
	p.mu.Lock()
	defer p.mu.Unlock()

	key, ok := p.bySession[sessionID]
	if !ok {
		return nil
	}

	entry, exists := p.byKey[key]
	if !exists {
		delete(p.bySession, sessionID)
		return nil
	}

	delete(p.byKey, key)
	delete(p.bySession, sessionID)

	return []PresenceEntry{entry}
}

func (p *PresenceStore) Snapshot() []PresenceEntry {
	p.mu.RLock()
	defer p.mu.RUnlock()

	out := make([]PresenceEntry, 0, len(p.byKey))
	for _, e := range p.byKey {
		out = append(out, e)
	}

	return out
}

type Broadcaster interface {
	BroadcastEvent(evt Event)
}
