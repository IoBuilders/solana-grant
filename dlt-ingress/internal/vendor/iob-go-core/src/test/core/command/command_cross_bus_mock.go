package commandbusmock

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
)

type CrossBusMock struct {
	mock.Mock
}

func (m *CrossBusMock) RegisterPort(port any) error {
	m.Called(port)
	return nil
}

func (m *CrossBusMock) Dispatch(
	ctx context.Context,
	cmd command.CrossCommand,
) (command.CrossResponse, error) {
	args := m.Called(ctx, cmd)

	var response command.CrossResponse
	if args.Get(0) != nil {
		response = args.Get(0).(command.CrossResponse)
	}

	return response, args.Error(1)
}
