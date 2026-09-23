package getfailedtransactions

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/transaction/failedtransaction"
	"dlt-ingress/src/main/dltingress/port/repository/transaction/failedtransaction"
)

type QueryHandler struct {
	failedTransactionRepository failedtransactionrepo.Repository
}

func NewHandler(failedTransactionRepository failedtransactionrepo.Repository) *QueryHandler {
	return &QueryHandler{failedTransactionRepository: failedTransactionRepository}
}

func (h *QueryHandler) Execute(ctx context.Context, query Query) (Response, error) {
	entities, err := h.failedTransactionRepository.FindByStatusPaginated(ctx, failedtransaction.StatusNotRetried, query.PaginationParams)
	if err != nil {
		return Response{}, err
	}

	total, err := h.failedTransactionRepository.CountAllByFilters(ctx, failedtransaction.StatusNotRetried)
	if err != nil {
		return Response{}, err
	}

	items := make([]*FailedTransactionQueryResponse, len(entities))
	for i, n := range entities {
		items[i] = ToItem(n)
	}

	return Response{
		Items:            items,
		TotalElements:    total,
		PaginationParams: query.PaginationParams,
	}, nil
}

func ToItem(ec failedtransaction.FailedTransaction) *FailedTransactionQueryResponse {
	return &FailedTransactionQueryResponse{
		TxId:         ec.TxId,
		NetworkId:    ec.NetworkId,
		CreatedAt:    ec.CreatedAt,
		ErrorDetails: ec.ErrorDetails,
	}
}
