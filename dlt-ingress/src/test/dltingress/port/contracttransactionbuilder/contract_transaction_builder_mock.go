package contracttransactionbuilder

import (
	"dlt-ingress/src/main/dltingress/port/contracttransactionbuilder"
	"dlt-ingress/src/main/dltingress/port/portcommon"

	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type ContractTransactionBuilderMock struct {
	mock.Mock
}

func (r *ContractTransactionBuilderMock) BuildTransaction(request *contracttransactionbuilder.BuildTransactionRequest) (*portcommon.TransactionResponse, error) {
	args := r.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*portcommon.TransactionResponse), args.Error(1)
}

func (r *ContractTransactionBuilderMock) OverrideGasLimit(originalRequest contracttransactionbuilder.BuildTransactionRequest, originalResponse portcommon.TransactionResponse, newGasLimit *amount.Amount) (*portcommon.TransactionResponse, error) {
	args := r.Called(originalRequest, originalResponse, newGasLimit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*portcommon.TransactionResponse), args.Error(1)
}

var _ contracttransactionbuilder.Port = (*ContractTransactionBuilderMock)(nil)
