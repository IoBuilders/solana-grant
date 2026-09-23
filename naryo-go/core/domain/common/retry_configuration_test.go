//go:build test

package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestRetryConfiguration_Default(t *testing.T) {
	retry := DefaultRetryConfiguration()
	assert.Equal(t, 5, retry.MaxRetries)
	assert.Equal(t, time.Second, retry.InitialDelay)
	assert.Equal(t, 30*time.Second, retry.MaxDelay)
	assert.Equal(t, 2.0, retry.Multiplier)
}

func TestRetryConfiguration_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		retry, err := NewRetryConfiguration(3, 500*time.Millisecond, 10*time.Second, 1.5)
		assert.NoError(t, err)
		assert.Equal(t, 3, retry.MaxRetries)
		assert.Equal(t, 500*time.Millisecond, retry.InitialDelay)
		assert.Equal(t, 10*time.Second, retry.MaxDelay)
		assert.Equal(t, 1.5, retry.Multiplier)
	})

	t.Run("NegativeMaxRetries", func(t *testing.T) {
		_, err := NewRetryConfiguration(-1, time.Second, 30*time.Second, 2.0)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("ZeroInitialDelay", func(t *testing.T) {
		_, err := NewRetryConfiguration(5, 0, 30*time.Second, 2.0)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("MaxDelayBelowInitialDelay", func(t *testing.T) {
		_, err := NewRetryConfiguration(5, time.Second, 500*time.Millisecond, 2.0)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("MultiplierBelowOne", func(t *testing.T) {
		_, err := NewRetryConfiguration(5, time.Second, 30*time.Second, 0.5)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
