package testevent

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type EventBusMock struct {
	mock.Mock
}

func (m *EventBusMock) Publish(ctx context.Context, event event.Event) error {
	args := m.Called(ctx, event)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

var _ event.Bus = (*EventBusMock)(nil)
