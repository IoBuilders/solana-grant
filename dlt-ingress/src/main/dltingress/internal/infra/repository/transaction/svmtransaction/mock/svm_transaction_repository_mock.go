package mocksvmtransactionrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/svmtransaction"
)

func (m *Postgres) FindByTxId(ctx context.Context, txId string) (*svmtransaction.SvmTransaction, error) {
	args := m.Called(ctx, txId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*svmtransaction.SvmTransaction), args.Error(1)
}

func (m *Postgres) HardDelete(ctx context.Context, entity *svmtransaction.SvmTransaction) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}
