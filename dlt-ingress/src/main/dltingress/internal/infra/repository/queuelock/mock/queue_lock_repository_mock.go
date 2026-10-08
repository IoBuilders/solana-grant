package mockqueuelockrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/queuelock"
)

func (m *Postgres) FindAndLockByNetworkId(ctx context.Context, networkId string) (*queuelock.QueueLock, error) {
	args := m.Called(ctx, networkId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*queuelock.QueueLock), nil
}

func (m *Postgres) CreateIfNotExists(ctx context.Context, entity *queuelock.QueueLock) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}
