package transitfailedtransactiontoretried

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/transaction/failedtransaction"
	"dlt-ingress/src/main/dltingress/port/repository/transaction/failedtransaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type Handler struct {
	eventBus                    event.Bus
	failedTransactionRepository failedtransactionrepo.Repository
}

func NewHandler(
	eventBus event.Bus,
	failedTransactionRepository failedtransactionrepo.Repository,
) *Handler {
	return &Handler{
		eventBus:                    eventBus,
		failedTransactionRepository: failedTransactionRepository,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd *Command) (*Response, error) {
	failedTransaction, err := h.failedTransactionRepository.FindByTxId(ctx, cmd.TxId)
	if err != nil {
		return nil, err
	}

	if err = failedTransaction.TransitToRetried(); err != nil {
		return nil, err
	}

	err = h.failedTransactionRepository.Save(ctx, failedTransaction)
	if err != nil {
		return nil, err
	}

	evt := failedtransaction.TransitFailedTransactionToRetriedEvent{
		BaseEvent: *event.NewBaseEvent(),
		TxId:      failedTransaction.TxId,
		NetworkId: failedTransaction.NetworkId,
	}
	err = h.eventBus.Publish(ctx, evt)
	if err != nil {
		return nil, err
	}

	return &Response{evt}, nil
}
