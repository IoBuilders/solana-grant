package cache

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type CachePortMock struct {
	mock.Mock
}

func (m *CachePortMock) Get(ctx context.Context, key string) (any, error) {
	args := m.Called(ctx, key)
	if args.Get(0) != nil {
		return args.Get(0), nil
	}
	return nil, args.Error(1)
}

func (m *CachePortMock) Set(ctx context.Context, key string, value any) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *CachePortMock) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}
