package node

import (
	"time"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// BlockSubscriptionMethod discriminates how new blocks are obtained from a Node.
type BlockSubscriptionMethod string

const (
	BlockSubscriptionMethodPubSub BlockSubscriptionMethod = "PUBSUB"
	BlockSubscriptionMethodPoll   BlockSubscriptionMethod = "POLL"
)

func (m BlockSubscriptionMethod) IsValid() bool {
	switch m {
	case BlockSubscriptionMethodPubSub, BlockSubscriptionMethodPoll:
		return true
	}
	return false
}

func (m BlockSubscriptionMethod) String() string {
	return string(m)
}

// BlockSubscriptionMethodConfiguration carries the settings specific to one
// BlockSubscriptionMethod.
type BlockSubscriptionMethodConfiguration interface {
	Method() BlockSubscriptionMethod
}

// PollBlockSubscriptionMethodConfiguration subscribes by polling the node at
// a fixed interval.
type PollBlockSubscriptionMethodConfiguration struct {
	Interval time.Duration
}

func NewPollBlockSubscriptionMethodConfiguration(interval time.Duration) (PollBlockSubscriptionMethodConfiguration, error) {
	if interval <= 0 {
		return PollBlockSubscriptionMethodConfiguration{}, domainerrors.NewInvalidFieldError(
			"Interval", "PollBlockSubscriptionMethodConfiguration", "must be > 0",
		)
	}
	return PollBlockSubscriptionMethodConfiguration{Interval: interval}, nil
}

func (PollBlockSubscriptionMethodConfiguration) Method() BlockSubscriptionMethod {
	return BlockSubscriptionMethodPoll
}

// PubSubBlockSubscriptionMethodConfiguration subscribes through the node's
// native publish/subscribe notifications.
type PubSubBlockSubscriptionMethodConfiguration struct{}

func NewPubSubBlockSubscriptionMethodConfiguration() PubSubBlockSubscriptionMethodConfiguration {
	return PubSubBlockSubscriptionMethodConfiguration{}
}

func (PubSubBlockSubscriptionMethodConfiguration) Method() BlockSubscriptionMethod {
	return BlockSubscriptionMethodPubSub
}

// BlockSubscriptionConfiguration describes how a Node ingests new blocks,
// starting from InitialSlot.
type BlockSubscriptionConfiguration struct {
	MethodConfiguration BlockSubscriptionMethodConfiguration
	InitialSlot         uint64
}

func NewBlockSubscriptionConfiguration(methodConfiguration BlockSubscriptionMethodConfiguration, initialSlot uint64) (BlockSubscriptionConfiguration, error) {
	if methodConfiguration == nil {
		return BlockSubscriptionConfiguration{}, domainerrors.NewEmptyFieldError(
			"MethodConfiguration", "BlockSubscriptionConfiguration",
		)
	}
	return BlockSubscriptionConfiguration{
		MethodConfiguration: methodConfiguration,
		InitialSlot:         initialSlot,
	}, nil
}
