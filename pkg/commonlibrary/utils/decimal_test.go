package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFloatPtrToNullDecimalAndBack(t *testing.T) {
	t.Run("nil round trips to nil", func(t *testing.T) {
		got, err := NullDecimalToFloatPtr(FloatPtrToNullDecimal(nil))
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("value round trips", func(t *testing.T) {
		secs := 12.34
		got, err := NullDecimalToFloatPtr(FloatPtrToNullDecimal(&secs))
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.InDelta(t, secs, *got, 0.001)
	})

	t.Run("zero NullDecimal returns nil", func(t *testing.T) {
		got, err := NullDecimalToFloatPtr(FloatPtrToNullDecimal(nil))
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
