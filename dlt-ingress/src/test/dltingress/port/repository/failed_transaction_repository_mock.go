package mocks

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/transaction/failedtransaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
)

func (m *FailedTransactionRepositoryMock) FindByTxId(ctx context.Context, txId string) (*failedtransaction.FailedTransaction, error) {
	args := m.Called(ctx, txId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*failedtransaction.FailedTransaction), args.Error(1)
}

func (m *FailedTransactionRepositoryMock) FindByStatusPaginated(ctx context.Context, status failedtransaction.Status, params pagination.PaginationParams) ([]failedtransaction.FailedTransaction, error) {
	args := m.Called(ctx, status, params)
	return args.Get(0).([]failedtransaction.FailedTransaction), args.Error(1)
}

func (m *FailedTransactionRepositoryMock) CountAllByFilters(ctx context.Context, status failedtransaction.Status) (int, error) {
	args := m.Called(ctx, status)
	return args.Int(0), args.Error(1)
}

func (m *FailedTransactionRepositoryMock) ExistByTxIdAndStatus(ctx context.Context, txId string, status failedtransaction.Status) (bool, error) {
	args := m.Called(ctx, txId, status)
	return args.Bool(0), args.Error(1)
}
