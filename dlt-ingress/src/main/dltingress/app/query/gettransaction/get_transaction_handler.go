package gettransaction

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/common"
	"dlt-ingress/src/main/dltingress/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/port/repository/transaction/evmtransaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
)

type QueryHandler struct {
	evmRepo evmtransactionrepo.Repository
}

type Response struct {
	TxId       string
	NetworkId  string
	NetworkUrl string
	Dlt        string
	Evm        *EvmTransactionModel
}

type EvmTransactionModel struct {
	FromAddress          string
	ToAddress            string
	Nonce                *amount.Amount
	Value                *amount.Amount
	TransactionType      uint
	GasLimit             *amount.Amount
	GasPrice             *amount.Amount
	MaxPriorityFeePerGas *amount.Amount
	MaxFeePerGas         *amount.Amount
	Data                 string
}

func NewHandler(evmRepo evmtransactionrepo.Repository) *QueryHandler {
	return &QueryHandler{evmRepo: evmRepo}
}

func (h *QueryHandler) Execute(ctx context.Context, query Query) (query.Response, error) {

	dlt, err := common.ParseDlt(query.Dlt)
	if err != nil {
		return nil, err
	}

	switch dlt {
	case common.EVM:
		tx, err := h.evmRepo.FindByTxId(ctx, query.TxId)
		if err != nil {
			return nil, err
		}
		return Response{
			TxId:       tx.TxId,
			NetworkId:  tx.NetworkId,
			NetworkUrl: tx.NetworkUrl,
			Dlt:        string(tx.Dlt),
			Evm: &EvmTransactionModel{
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
