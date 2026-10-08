package gettransaction

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/query"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/evmtransaction"
)

type QueryHandler struct {
	evmRepo evmtransaction.Repository
}

func NewHandler(evmRepo evmtransaction.Repository) *QueryHandler {
	return &QueryHandler{evmRepo: evmRepo}
}

func (h *QueryHandler) Execute(ctx context.Context, q Query) (any, error) {

	dlt, err := common.ParseDlt(q.Dlt)
	if err != nil {
		return nil, err
	}

	switch dlt {
	case common.EVM:
		tx, err := h.evmRepo.FindByTxId(ctx, q.TxId)
		if err != nil {
			return nil, err
		}
		return Response{
			TxId:       tx.TxId,
			NetworkId:  tx.NetworkId,
			NetworkUrl: tx.NetworkUrl,
			Dlt:        string(tx.Dlt),
			Evm: &query.EvmTransaction{
				FromAddress:          tx.FromAddress,
				ToAddress:            tx.ToAddress,
				Nonce:                tx.Nonce,
				Value:                tx.Value,
				TransactionType:      tx.TransactionType,
				GasLimit:             tx.GasLimit,
				GasPrice:             tx.GasPrice,
				MaxPriorityFeePerGas: tx.MaxPriorityFeePerGas,
				MaxFeePerGas:         tx.MaxFeePerGas,
				Data:                 tx.Data,
			},
		}, nil
	default:
		return nil, domainerrors.NewInvalidDltDomainError(string(dlt))
	}
}
