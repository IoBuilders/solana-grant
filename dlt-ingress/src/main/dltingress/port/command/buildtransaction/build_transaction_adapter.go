package buildtransactioncross

import (
	"context"
	"dlt-ingress/src/main/dltingress/app/command/buildtransaction"
	"dlt-ingress/src/main/dltingress/port/event/buildtransaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
)

type CrossCommandAdapter struct {
	commandBus command.Bus
}

func NewCrossCommandAdapter(commandBus command.Bus) *CrossCommandAdapter {
	return &CrossCommandAdapter{commandBus: commandBus}
}

func (h *CrossCommandAdapter) Execute(ctx context.Context, cmd *CrossCommand) (*CrossResponse, error) {
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

	return &CrossResponse{
		TransactionBuiltCrossEvent: buildtransactionevents.TransactionBuiltCrossEvent{
			BaseEvent:    resp.BaseEvent,
			DltAccountId: resp.DltAccountId,
			Payload:      resp.Payload,
		},
	}, nil
}
