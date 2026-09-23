package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type BlockhashProviderMock struct {
	mock.Mock
}

func (m *BlockhashProviderMock) GetRecentBlockhash(ctx context.Context, networkId string) (string, error) {
	args := m.Called(ctx, networkId)
	return args.String(0), args.Error(1)
}
