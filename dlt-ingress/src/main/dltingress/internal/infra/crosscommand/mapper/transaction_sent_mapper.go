package mapper

import (
	"dlt-ingress/src/main/dltingress/internal/app/command/signandsend"
	"dlt-ingress/src/main/dltingress/port/crossevent/signandsend"
)

func ToCrossEvent(resp *signandsend.Response) signandsendevents.TransactionSentCrossEvent {
	crossEvent := signandsendevents.TransactionSentCrossEvent{
		BaseEvent: resp.BaseEvent,
		TxId:      resp.TxId,
		Dlt:       resp.Dlt,
	}

	if resp.EvmTransactionEventModel != nil {
		crossEvent.EvmTransactionCrossEventModel = &signandsendevents.EvmTransactionCrossEventModel{
			FromAddress:          resp.EvmTransactionEventModel.FromAddress,
			ToAddress:            resp.EvmTransactionEventModel.ToAddress,
			Nonce:                resp.EvmTransactionEventModel.Nonce,
			Value:                resp.EvmTransactionEventModel.Value,
			Data:                 resp.EvmTransactionEventModel.Data,
			TransactionType:      resp.EvmTransactionEventModel.TransactionType,
			GasLimit:             resp.EvmTransactionEventModel.GasLimit,
			GasPrice:             resp.EvmTransactionEventModel.GasPrice,
			MaxPriorityFeePerGas: resp.EvmTransactionEventModel.MaxPriorityFeePerGas,
			MaxFeePerGas:         resp.EvmTransactionEventModel.MaxFeePerGas,
		}
	}

	if resp.SvmTransactionEventModel != nil {
		crossEvent.SvmTransactionCrossEventModel = &signandsendevents.SvmTransactionCrossEventModel{
			FeePayer:              resp.SvmTransactionEventModel.FeePayer,
			RecentBlockhash:       resp.SvmTransactionEventModel.RecentBlockhash,
			SerializedTransaction: resp.SvmTransactionEventModel.SerializedTransaction,
			CuLimit:               resp.SvmTransactionEventModel.CuLimit,
			CuPrice:               resp.SvmTransactionEventModel.CuPrice,
		}
	}

	return crossEvent
}
