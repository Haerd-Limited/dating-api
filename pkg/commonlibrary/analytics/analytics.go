package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Event represents a single tracked action.
type Event struct {
	Name      string
	UserID    *string
	SessionID *string
	Props     map[string]any
	Occurred  time.Time
}

// Emitter writes events to Postgres using a buffered worker.
type Emitter struct {
	db     *sqlx.DB
	logger *zap.Logger

	ch      chan Event
	wg      sync.WaitGroup
	closing chan struct{}
}

// Default is the process-wide emitter. Initialize with InitDefault.
var Default *Emitter

// OptOutFunc, if set, determines if a user has opted out of analytics.
var OptOutFunc func(userID string) bool

// InitDefault creates and starts the default emitter.
func InitDefault(db *sqlx.DB, logger *zap.Logger, buffer int) {
	Default = NewEmitter(db, logger, buffer)
	Default.Start()
}

// ShutdownDefault stops the default emitter if initialized.
func ShutdownDefault() {
	if Default != nil {
		Default.Stop()
	}
}

func NewEmitter(db *sqlx.DB, logger *zap.Logger, buffer int) *Emitter {
	if buffer <= 0 {
		buffer = 1024
	}

	return &Emitter{
		db:      db,
		logger:  logger,
		ch:      make(chan Event, buffer),
		closing: make(chan struct{}),
	}
}

// Start begins the background worker.
func (e *Emitter) Start() {
	e.wg.Add(1)

	go func() {
		defer e.wg.Done()

		for {
			select {
			case ev := <-e.ch:
				e.insert(context.Background(), ev)
			case <-e.closing:
				// drain channel
				for {
					select {
					case ev := <-e.ch:
						e.insert(context.Background(), ev)
					default:
						return
					}
				}
			}
		}
	}()
}

// Stop gracefully stops the worker.
func (e *Emitter) Stop() {
	close(e.closing)
	e.wg.Wait()
}

// Track queues an event. If the buffer is full, it falls back to a best-effort synchronous write.
func (e *Emitter) Track(ctx context.Context, name string, userID *string, sessionID *string, props map[string]any) {
	if userID != nil && OptOutFunc != nil && OptOutFunc(*userID) {
		return
	}

	ev := Event{
		Name:      name,
		UserID:    userID,
		SessionID: sessionID,
		Props:     props,
		Occurred:  time.Now().UTC(),
	}
	select {
	case e.ch <- ev:
	default:
		// buffer full, write synchronously but don't fail callers
		e.insert(ctx, ev)
	}
}

// Track is a convenience wrapper for the Default emitter.
func Track(ctx context.Context, name string, userID *string, sessionID *string, props map[string]any) {
	if Default == nil {
		return
	}

	Default.Track(ctx, name, userID, sessionID, props)
}

func (e *Emitter) insert(ctx context.Context, ev Event) {
	// Use sql.Named for clarity and to avoid props json casting issues.
	_, err := e.db.ExecContext(
		ctx,
		`INSERT INTO events (occurred_at, user_id, session_id, name, props, version)
         VALUES ($1, $2, $3, $4, $5::jsonb, 1)`,
		ev.Occurred,
		nullUUID(ev.UserID),
		ev.SessionID,
		ev.Name,
		mustJSON(ev.Props),
	)
	if err != nil && e.logger != nil {
		e.logger.Sugar().Warnf("analytics insert failed: %v", err)
	}
}

// Helper to convert optional user id into nullable type that works with Exec.
func nullUUID(s *string) any {
	if s == nil {
		return sql.NullString{Valid: false}
	}

	return *s
}

// Marshal to JSON string with minimal overhead; props may be nil.
func mustJSON(m map[string]any) string {
	if m == nil {
		return "{}"
	}
	// Lightweight encoding to avoid importing a second dependency.
	// We rely on sql to cast string to jsonb; this is safe for pre-validated maps.
	// For robustness, we could use encoding/json; cost is acceptable.
	b, _ := jsonMarshal(m)

	return string(b)
}

// tiny shim to keep imports lean; we keep it separate for easy replacement in tests.
var jsonMarshal = func(v any) ([]byte, error) {
	return json.Marshal(v)
}
