package mockevmtransactionrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/evmtransaction"
)

func (m *Postgres) FindByTxId(ctx context.Context, txId string) (*evmtransaction.EvmTransaction, error) {
	args := m.Called(ctx, txId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*evmtransaction.EvmTransaction), args.Error(1)
}

func (m *Postgres) HardDelete(ctx context.Context, entity *evmtransaction.EvmTransaction) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}
