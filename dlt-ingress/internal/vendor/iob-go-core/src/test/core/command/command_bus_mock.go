package commandbusmock

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"

	"github.com/stretchr/testify/mock"
)

type CommandBusMock struct {
	mock.Mock
}

func (m *CommandBusMock) RegisterHandler(handler any) error {
	m.Called(handler)
	return nil
}

func (m *CommandBusMock) Dispatch(ctx context.Context, cmd command.Command) (command.Response, error) {
	args := m.Called(ctx, cmd)
	var resp command.Response
	if args.Get(0) != nil {
		resp = args.Get(0).(command.Response)
	}
	return resp, args.Error(1)
}
