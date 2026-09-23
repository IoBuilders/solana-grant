//go:build test

package broadcaster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestNewExchange_Valid(t *testing.T) {
	e, err := NewExchange("Naryo-Events")

	require.NoError(t, err)
	assert.Equal(t, "Naryo-Events", e.String())
}

func TestNewExchange_Empty(t *testing.T) {
	_, err := NewExchange(" ")

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
}
