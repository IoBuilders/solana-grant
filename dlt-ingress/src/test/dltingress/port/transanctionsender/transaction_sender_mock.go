package transanctionsender

import (
	"context"
	"dlt-ingress/src/main/dltingress/port/transanctionsender"

	"github.com/stretchr/testify/mock"
)

type TransactionSenderMock struct {
	mock.Mock
}

func (m *TransactionSenderMock) SendTransaction(ctx context.Context, request transanctionsender.SendTransactionRequest) (*transanctionsender.SendTransactionResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*transanctionsender.SendTransactionResponse), args.Error(1)
}

var _ transanctionsender.Port = (*TransactionSenderMock)(nil)
