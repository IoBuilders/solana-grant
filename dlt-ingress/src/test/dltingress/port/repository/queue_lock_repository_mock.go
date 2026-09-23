package mocks

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/queuelock"
)

func (m *QueueLockRepositoryMock) FindAndLockByNetworkId(ctx context.Context, networkId string) (*queuelock.QueueLock, error) {
	args := m.Called(ctx, networkId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*queuelock.QueueLock), nil
}

func (m *QueueLockRepositoryMock) CreateIfNotExists(ctx context.Context, entity *queuelock.QueueLock) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}
