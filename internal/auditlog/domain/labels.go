package domain

import (
	"fmt"
	"strings"
	"time"
)

// ActionLabel returns a human-readable label for an audit entry.
func ActionLabel(method, path string) string {
	p := strings.ToLower(path)
	m := strings.ToUpper(method)

	switch {
	case m == "POST" && strings.Contains(p, "/session"):
		return "Admin login"
	case m == "DELETE" && strings.Contains(p, "/session"):
		return "Admin logout"
	case m == "POST" && strings.Contains(p, "/approve"):
		return "Approved video verification"
	case m == "POST" && strings.Contains(p, "/reject"):
		return "Rejected video verification"
	case m == "POST" && strings.Contains(p, "/resolve"):
		return "Resolved report"
	case m == "POST" && strings.Contains(p, "/broadcast"):
		return "Sent waitlist broadcast"
	default:
		return fmt.Sprintf("%s %s", method, path)
	}
}

// IsMeaningfulAction returns true for mutations worth showing on the Events page.
func IsMeaningfulAction(method, path string) bool {
	if method == "GET" || method == "OPTIONS" {
		return false
	}

	p := strings.ToLower(path)

	switch {
	case strings.Contains(p, "/approve"),
		strings.Contains(p, "/reject"),
		strings.Contains(p, "/resolve"),
		strings.Contains(p, "/broadcast"),
		strings.Contains(p, "/session"):
		return true
	default:
		return false
	}
}

type EventRow struct {
	ID         string
	Label      string
	ActorName  *string
	TargetID   *string
	Method     string
	Path       string
	StatusCode int
	OccurredAt time.Time
}

// ActionPathPattern returns a SQL LIKE pattern for filtering by action type query param.
func ActionPathPattern(action string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "approve":
		return "%/approve", true
	case "reject":
		return "%/reject", true
	case "resolve":
		return "%/resolve", true
	case "broadcast":
		return "%/broadcast", true
	case "session":
		return "%/session", true
	default:
		return "", false
	}
}

// ResourceTypeFromPath derives a resource type label from an audit log path.
func ResourceTypeFromPath(path string) string {
	p := strings.ToLower(path)

	switch {
	case strings.Contains(p, "/verification/"):
		return "video_verification"
	case strings.Contains(p, "/reports/"):
		return "report"
	case strings.Contains(p, "/broadcast"):
		return "broadcast"
	case strings.Contains(p, "/session"):
		return "session"
	default:
		return ""
	}
}

func EntryToEventRow(e Entry) EventRow {
	id := e.OccurredAt.UTC().Format(time.RFC3339Nano)
	if e.TargetID != nil && *e.TargetID != "" {
		id = *e.TargetID + "-" + id
	}

	return EventRow{
		ID:         id,
		Label:      ActionLabel(e.Method, e.Path),
		ActorName:  e.ActorName,
		TargetID:   e.TargetID,
		Method:     e.Method,
		Path:       e.Path,
		StatusCode: e.StatusCode,
		OccurredAt: e.OccurredAt,
	}
}
