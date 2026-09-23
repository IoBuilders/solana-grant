package boundedblockingqueue

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type BoundedBlockingQueueMock struct {
	mock.Mock
}

func (m *BoundedBlockingQueueMock) Take(ctx context.Context, networkId string) error {
	args := m.Called(ctx, networkId)
	return args.Error(0)
}

func (m *BoundedBlockingQueueMock) Put(ctx context.Context, networkId string, id uuid.UUID) error {
	args := m.Called(ctx, networkId, id)
	return args.Error(0)
}
