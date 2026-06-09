package matchslot

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestNotifierWithNilRepoDoesNotPanic(t *testing.T) {
	n := NewNotifier(zaptest.NewLogger(t), nil, nil, nil, nil)

	assert.NotPanics(t, func() {
		n.NotifySlotFreed(context.Background(), "user-1")
		n.NotifySlotFilled(context.Background(), "user-1")
	})
}
