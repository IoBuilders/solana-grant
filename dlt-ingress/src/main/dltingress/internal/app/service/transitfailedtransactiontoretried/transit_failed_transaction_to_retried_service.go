package transitfailedtransactiontoretried

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/command/transitfailedtransactiontoretried"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/failedtransaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

type AppServiceInterface interface {
	Execute(ctx context.Context, request *Request) error
}

type AppService struct {
	commmandBus                 command.Bus
	failedTransactionRepository failedtransaction.Repository
}

func NewAppService(commmandBus command.Bus, failedTransactionRepository failedtransaction.Repository) *AppService {
	return &AppService{commmandBus: commmandBus, failedTransactionRepository: failedTransactionRepository}
}

func (s *AppService) Execute(ctx context.Context, request *Request) error {
	exists, err := s.failedTransactionRepository.ExistByTxIdAndStatus(ctx, request.TxId, failedtransaction.StatusNotRetried)
	if err != nil {
		return err
	}
	if exists {
		_, err := s.commmandBus.Dispatch(ctx, &transitfailedtransactiontoretried.Command{
			TxId: request.TxId,
		})
		return err
	} else {
		logger.WarnWithCtx(ctx, "Skipping transaction status transition because it does not exist", "txId", request.TxId)
		return nil
	}
}
