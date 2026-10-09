package mockcustodykeyrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
)

func (m *Postgres) FindByDltAccountId(ctx context.Context, dltAccountId string) (*custodykey.CustodyKey, error) {
	args := m.Called(ctx, dltAccountId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*custodykey.CustodyKey), args.Error(1)
}

func (m *Postgres) FindByDltAccountIds(ctx context.Context, dltAccountIds []string) ([]*custodykey.CustodyKey, error) {
	args := m.Called(ctx, dltAccountIds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*custodykey.CustodyKey), args.Error(1)
}
