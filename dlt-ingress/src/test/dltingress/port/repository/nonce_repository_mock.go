package mocks

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/nonce"
)

func (m *NonceRepositoryMock) FindAndLockByDltAccountIdAndNetworkId(ctx context.Context, dltAccountId string, networkId string) (*nonce.Nonce, error) {
	args := m.Called(ctx, dltAccountId, networkId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*nonce.Nonce), args.Error(1)
}

func (m *NonceRepositoryMock) DeleteAll(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return nil
}
