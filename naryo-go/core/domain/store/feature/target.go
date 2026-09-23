package feature

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// TargetType is the kind of event an EventTarget routes.
type TargetType string

const (
	TargetTypeBlock         TargetType = "BLOCK"
	TargetTypeTransaction   TargetType = "TRANSACTION"
	TargetTypeContractEvent TargetType = "CONTRACT_EVENT"
)

func (t TargetType) IsValid() bool {
	switch t {
	case TargetTypeBlock, TargetTypeTransaction, TargetTypeContractEvent:
		return true
	}
	return false
}

func (t TargetType) String() string {
	return string(t)
}

type EventTarget struct {
	targetType  TargetType
	destination Destination
}

func NewEventTarget(targetType TargetType, destination Destination) (EventTarget, error) {
	t := EventTarget{targetType: targetType, destination: destination}
	if err := t.Validate(); err != nil {
		return EventTarget{}, err
	}
	return t, nil
}

func (t EventTarget) TargetType() TargetType {
	return t.targetType
}

func (t EventTarget) Destination() Destination {
	return t.destination
}

func (t EventTarget) Validate() error {
	if !t.targetType.IsValid() {
		return domainerrors.NewInvalidFieldError("TargetType", "EventTarget", "unknown target type "+t.targetType.String())
	}
	return nil
}
