//go:build test

package eventmapper

import (
	"github.com/stretchr/testify/mock"

	coremapping "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/mapping"
)

// EventToJsonMapperMock is a testify mock double for
// coremapping.Mapper[BlockchainEventSource, []byte], so any module's
// producer tests can mock the shared mapper instead of hand-rolling their
// own test double.
type EventToJsonMapperMock struct {
	mock.Mock
}

func (m *EventToJsonMapperMock) Map(source BlockchainEventSource) ([]byte, error) {
	args := m.Called(source)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

var _ coremapping.Mapper[BlockchainEventSource, []byte] = (*EventToJsonMapperMock)(nil)
