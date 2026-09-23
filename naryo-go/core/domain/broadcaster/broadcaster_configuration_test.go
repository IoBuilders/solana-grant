//go:build test

package broadcaster

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestGenericConfiguration_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		id := uuid.New()
		additionalProperties := map[string]interface{}{"foo": "bar"}

		c, err := NewGenericConfiguration(id, Type("KAFKA"), additionalProperties)
		assert.NoError(t, err)
		assert.Equal(t, id, c.ID())
		assert.Equal(t, Type("KAFKA"), c.Type())
		assert.Equal(t, additionalProperties, c.AdditionalProperties())
	})

	t.Run("NilID", func(t *testing.T) {
		_, err := NewGenericConfiguration(uuid.Nil, Type("KAFKA"), nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("EmptyType", func(t *testing.T) {
		_, err := NewGenericConfiguration(uuid.New(), Type(""), nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
