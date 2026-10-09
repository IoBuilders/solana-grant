package buildtransactionadapter

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/command/buildtransaction"
	"dlt-ingress/src/main/dltingress/port/crosscommand/buildtransaction"
	"dlt-ingress/src/main/dltingress/port/crossevent/buildtransaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
)

type CrossCommandAdapter struct {
	commandBus command.Bus
}

func NewCrossCommandAdapter(commandBus command.Bus) *CrossCommandAdapter {
	return &CrossCommandAdapter{commandBus: commandBus}
}

func (h *CrossCommandAdapter) Execute(ctx context.Context, cmd *buildtransactioncross.CrossCommand) (*buildtransactioncross.CrossResponse, error) {
	result, err := h.commandBus.Dispatch(ctx, &buildtransaction.Command{
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

	resp := result.(*buildtransaction.Response)

	return &buildtransactioncross.CrossResponse{
		TransactionBuiltCrossEvent: buildtransactionevents.TransactionBuiltCrossEvent{
			BaseEvent:    resp.BaseEvent,
			DltAccountId: resp.DltAccountId,
			Payload:      resp.Payload,
		},
	}, nil
}
