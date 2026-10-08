package signandsendadapter

import (
	"context"

	"dlt-ingress/src/main/dltingress/internal/app/command/signandsend"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/infra/crosscommand/mapper"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
	"dlt-ingress/src/main/dltingress/port/crosscommand/signandsend"
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

func (h *CrossCommandAdapter) Execute(ctx context.Context, cmd *signandsendcross.CrossCommand) (*signandsendcross.CrossResponse, error) {
	if err := rejectNestedCalls(cmd.MethodArgs); err != nil {
		return nil, err
	}

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

	return &signandsendcross.CrossResponse{TransactionSentCrossEvent: mapper.ToCrossEvent(resp)}, nil
}

func rejectNestedCalls(methodArgs map[string]any) error {
	for argName, value := range methodArgs {
		if _, nested := value.([]portcommon.Invocation); nested {
			return domainerrors.NewNestedCallsNotSupportedDomainError(argName)
		}
	}
	return nil
}
