package mocktxqueueslotrepo

import "context"

func (m *Postgres) CountByNetworkIdAndNotExpired(ctx context.Context, networkId string) (int, error) {
	args := m.Called(ctx, networkId)
	return args.Int(0), args.Error(1)
}

func (m *Postgres) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *Postgres) DeleteFirstByNetworkId(ctx context.Context, networkId string) error {
	args := m.Called(ctx, networkId)
	return args.Error(0)
}
