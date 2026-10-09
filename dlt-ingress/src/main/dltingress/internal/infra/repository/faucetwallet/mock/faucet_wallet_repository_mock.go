package mockfaucetwalletrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/faucetwallet"
)

func (m *Postgres) FindByNetworkId(ctx context.Context, networkId string) (*faucetwallet.FaucetWallet, error) {
	args := m.Called(ctx, networkId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*faucetwallet.FaucetWallet), args.Error(1)
}

func (m *Postgres) ExistByNetworkId(ctx context.Context, networkId string) (bool, error) {
	args := m.Called(ctx, networkId)
	return args.Bool(0), args.Error(1)
}
