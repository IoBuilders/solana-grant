package mocks

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/transaction/evmtransaction"
)

func (m *EvmTransactionRepositoryMock) FindByTxId(ctx context.Context, txId string) (*evmtransaction.EvmTransaction, error) {
	args := m.Called(ctx, txId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*evmtransaction.EvmTransaction), args.Error(1)
}
