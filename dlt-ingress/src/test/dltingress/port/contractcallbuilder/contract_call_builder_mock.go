//go:build test

package contractcallbuilder

import (
	"dlt-ingress/src/main/dltingress/port/contractcallbuilder"

	"github.com/stretchr/testify/mock"
)

type ContractCallBuilderMock struct {
	mock.Mock
}

func (m *ContractCallBuilderMock) BuildCall(request *contractcallbuilder.BuildCallRequest) (*contractcallbuilder.CallData, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contractcallbuilder.CallData), args.Error(1)
}

func (m *ContractCallBuilderMock) DecodeResult(methodName string, result string) (map[string]any, error) {
	args := m.Called(methodName, result)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

var _ contractcallbuilder.Port = (*ContractCallBuilderMock)(nil)
