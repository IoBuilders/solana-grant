package savefailedtransaction

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/transaction/failedtransaction"
	"dlt-ingress/src/main/dltingress/port/repository/transaction/failedtransaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type CommandHandler struct {
	failedTransactionRepo failedtransactionrepo.Repository
	eventBus              event.Bus
}

func NewCommandHandler(failedTransactionRepo failedtransactionrepo.Repository, eventBus event.Bus) *CommandHandler {
	return &CommandHandler{failedTransactionRepo, eventBus}
}

func (h *CommandHandler) Handle(ctx context.Context, cmd *Command) (*Response, error) {
	failedTransaction := failedtransaction.NewFailedTransaction(cmd.TxId, cmd.NetworkId, cmd.ErrorDetails)
	if err := h.failedTransactionRepo.Create(ctx, failedTransaction); err != nil {
		return nil, err
	}

	evt := failedtransaction.SavedEvent{
		BaseEvent:    *event.NewBaseEvent(),
		TxId:         failedTransaction.TxId,
		NetworkId:    failedTransaction.NetworkId,
		ErrorDetails: failedTransaction.ErrorDetails,
	}
	if err := h.eventBus.Publish(ctx, evt); err != nil {
		return nil, err
	}

	return &Response{evt}, nil
}
