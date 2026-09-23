package signandsendcross

import (
	"context"
	"dlt-ingress/src/main/dltingress/app/command/signandsend"
	"dlt-ingress/src/main/dltingress/port/event/signandsend"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
)

type CrossCommandAdapter struct {
	commandBus command.Bus
}

func NewCrossCommandAdapter(commandBus command.Bus) *CrossCommandAdapter {
	return &CrossCommandAdapter{
		commandBus: commandBus,
	}
}

func (h *CrossCommandAdapter) Execute(ctx context.Context, cmd *CrossCommand) (*CrossResponse, error) {
	result, err := h.commandBus.Dispatch(ctx, &signandsend.Command{
		SenderDltAccountId:   cmd.SenderDltAccountId,
		SignersDltAccountIds: cmd.SignersDltAccountIds,
		SmartContractId:      cmd.SmartContractId,
		SmartContractName:    cmd.SmartContractName,
		MethodName:           cmd.MethodName,
		MethodArgs:           cmd.MethodArgs,
		NetworkId:            cmd.NetworkId,
	})
	if err != nil {
		return nil, err
	}

	resp := result.(*signandsend.Response)

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

	return &CrossResponse{TransactionSentCrossEvent: crossEvent}, nil
}
