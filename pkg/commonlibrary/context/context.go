package context

import (
	"context"
)

type contextKey string

const UserIDKey contextKey = "userID"

const (
	AdminSessionIDKey   contextKey = "adminSessionID"
	AdminDisplayNameKey contextKey = "adminDisplayName"
)

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

func AdminActorFromContext(ctx context.Context) (sessionID, displayName string, ok bool) {
	sessionID, ok = ctx.Value(AdminSessionIDKey).(string)
	if !ok || sessionID == "" {
		return "", "", false
	}

	displayName, _ = ctx.Value(AdminDisplayNameKey).(string)

	return sessionID, displayName, true
}
