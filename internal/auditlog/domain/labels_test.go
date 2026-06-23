package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActionPathPattern(t *testing.T) {
	pattern, ok := ActionPathPattern("approve")
	assert.True(t, ok)
	assert.Equal(t, "%/approve", pattern)

	_, ok = ActionPathPattern("unknown")
	assert.False(t, ok)
}

func TestResourceTypeFromPath(t *testing.T) {
	assert.Equal(t, "video_verification", ResourceTypeFromPath("/admin/verification/video-attempts/abc/approve"))
	assert.Equal(t, "report", ResourceTypeFromPath("/admin/reports/xyz/resolve"))
	assert.Equal(t, "broadcast", ResourceTypeFromPath("/admin/waitlist/broadcast"))
	assert.Equal(t, "session", ResourceTypeFromPath("/admin/session"))
}
