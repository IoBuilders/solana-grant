package mocks

import (
	"dlt-ingress/src/main/dltingress/port/svm"

	"github.com/stretchr/testify/mock"
)

type SvmClientRegistryMock struct {
	mock.Mock
}

func (m *SvmClientRegistryMock) GetClientForNetworkId(networkId string) (svm.Client, error) {
	args := m.Called(networkId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(svm.Client), args.Error(1)
}

func (m *SvmClientRegistryMock) Shutdown() error {
	args := m.Called()
	return args.Error(0)
}
