//go:build test

package node

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestPollBlockSubscriptionMethodConfiguration_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		poll, err := NewPollBlockSubscriptionMethodConfiguration(time.Second)
		assert.NoError(t, err)
		assert.Equal(t, time.Second, poll.Interval)
		assert.Equal(t, BlockSubscriptionMethodPoll, poll.Method())
	})

	t.Run("ZeroInterval", func(t *testing.T) {
		_, err := NewPollBlockSubscriptionMethodConfiguration(0)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("NegativeInterval", func(t *testing.T) {
		_, err := NewPollBlockSubscriptionMethodConfiguration(-time.Second)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}

func TestPubSubBlockSubscriptionMethodConfiguration_New(t *testing.T) {
	pubsub := NewPubSubBlockSubscriptionMethodConfiguration()
	assert.Equal(t, BlockSubscriptionMethodPubSub, pubsub.Method())
}

func TestBlockSubscriptionConfiguration_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		subscription, err := NewBlockSubscriptionConfiguration(NewPubSubBlockSubscriptionMethodConfiguration(), 123)
		assert.NoError(t, err)
		assert.Equal(t, BlockSubscriptionMethodPubSub, subscription.MethodConfiguration.Method())
		assert.Equal(t, uint64(123), subscription.InitialSlot)
	})

	t.Run("NilMethodConfiguration", func(t *testing.T) {
		_, err := NewBlockSubscriptionConfiguration(nil, 0)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
