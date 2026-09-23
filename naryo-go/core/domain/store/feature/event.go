package feature

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// EventConfiguration enables persisting ingested events, and says which kinds
// of event go where.
type EventConfiguration struct {
	strategy Strategy
	targets  []EventTarget
}

func NewEventConfiguration(strategy Strategy, targets []EventTarget) (*EventConfiguration, error) {
	c := &EventConfiguration{strategy: strategy, targets: targets}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c EventConfiguration) Type() Type {
	return TypeEvent
}

func (c EventConfiguration) Strategy() Strategy {
	return c.strategy
}

func (c EventConfiguration) Targets() []EventTarget {
	return c.targets
}

func (c EventConfiguration) Target(targetType TargetType) (EventTarget, bool) {
	for _, t := range c.targets {
		if t.TargetType() == targetType {
			return t, true
		}
	}
	return EventTarget{}, false
}

func (c EventConfiguration) Validate() error {
	if !c.strategy.IsValid() {
		return domainerrors.NewInvalidFieldError("Strategy", "EventConfiguration", "unknown strategy "+c.strategy.String())
	}
	if len(c.targets) == 0 {
		return domainerrors.NewEmptyFieldError("Targets", "EventConfiguration")
	}
	seen := make(map[TargetType]struct{}, len(c.targets))
	for _, t := range c.targets {
		if err := t.Validate(); err != nil {
			return err
		}
		if _, duplicated := seen[t.TargetType()]; duplicated {
			return domainerrors.NewInvalidFieldError(
				"Targets", "EventConfiguration",
				"duplicate target type "+t.TargetType().String(),
			)
		}
		seen[t.TargetType()] = struct{}{}
	}
	return nil
}

var _ Configuration = (*EventConfiguration)(nil)
