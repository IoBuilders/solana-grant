package getfailedtransactions

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/query"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/failedtransaction"
)

type QueryHandler struct {
	failedTransactionRepository failedtransaction.Repository
}

func NewHandler(failedTransactionRepository failedtransaction.Repository) *QueryHandler {
	return &QueryHandler{failedTransactionRepository: failedTransactionRepository}
}

func (h *QueryHandler) Execute(ctx context.Context, q Query) (Response, error) {
	entities, err := h.failedTransactionRepository.FindByStatusPaginated(ctx, failedtransaction.StatusNotRetried, q.PaginationParams)
	if err != nil {
		return Response{}, err
	}

	total, err := h.failedTransactionRepository.CountAllByFilters(ctx, failedtransaction.StatusNotRetried)
	if err != nil {
		return Response{}, err
	}

	items := make([]*query.FailedTransaction, len(entities))
	for i, n := range entities {
		items[i] = &query.FailedTransaction{
			TxId:         n.TxId,
			NetworkId:    n.NetworkId,
			CreatedAt:    n.CreatedAt,
			ErrorDetails: n.ErrorDetails,
		}
	}

	return Response{
		Items:            items,
		TotalElements:    total,
		PaginationParams: q.PaginationParams,
	}, nil
}
