package mock

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/service/fundaccount"

	"github.com/stretchr/testify/mock"
)

type AppServiceMock struct {
	mock.Mock
}

func (m *AppServiceMock) Execute(ctx context.Context, request *fundaccount.Request) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

var _ fundaccount.AppServiceInterface = (*AppServiceMock)(nil)
