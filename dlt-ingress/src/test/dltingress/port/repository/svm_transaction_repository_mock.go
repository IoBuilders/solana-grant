package mocks

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/transaction/svmtransaction"
)

func (m *SvmTransactionRepositoryMock) FindByTxId(ctx context.Context, txId string) (*svmtransaction.SvmTransaction, error) {
	args := m.Called(ctx, txId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*svmtransaction.SvmTransaction), args.Error(1)
}
