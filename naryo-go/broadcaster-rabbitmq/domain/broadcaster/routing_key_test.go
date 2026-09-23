//go:build test

package broadcaster

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestNewRoutingKey_Valid(t *testing.T) {
	for _, value := range []string{"blocks", "naryo.blocks.42", "tx_events-v1.5abcXYZ"} {
		k, err := NewRoutingKey(value)

		require.NoError(t, err, value)
		assert.Equal(t, value, k.String())
	}
}

func TestNewRoutingKey_Empty(t *testing.T) {
	_, err := NewRoutingKey("  ")

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
}

func TestNewRoutingKey_TooLong(t *testing.T) {
	_, err := NewRoutingKey(strings.Repeat("a", 256))

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
	assert.ErrorContains(t, err, "255 bytes")
}

func TestNewRoutingKey_InvalidPattern(t *testing.T) {
	for _, value := range []string{"/events", "a..b", ".a", "a.", "a b", "a*", "a#"} {
		_, err := NewRoutingKey(value)

		assert.ErrorIs(t, err, domainerrors.ErrValidation, value)
	}
}
