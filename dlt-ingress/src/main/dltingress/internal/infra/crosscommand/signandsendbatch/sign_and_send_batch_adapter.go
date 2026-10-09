package signandsendbatchadapter

import (
	"context"

	"dlt-ingress/src/main/dltingress/internal/app/command/signandsend"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/infra/crosscommand/mapper"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
	"dlt-ingress/src/main/dltingress/port/crosscommand/signandsendbatch"
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

func (h *CrossCommandAdapter) Execute(ctx context.Context, cmd *signandsendbatchcross.CrossCommand) (*signandsendbatchcross.CrossResponse, error) {
	if err := requireCalls(cmd.Calls); err != nil {
		return nil, err
	}

	if cmd.Dispatch == nil {
		return nil, domainerrors.NewDispatchCallRequiredDomainError()
	}

	methodArgs, err := dispatchArgs(cmd.Dispatch, cmd.Calls)
	if err != nil {
		return nil, err
	}

	result, err := h.commandBus.Dispatch(ctx, &signandsend.Command{
		SenderDltAccountId:   cmd.SenderDltAccountId,
		SignersDltAccountIds: cmd.SignersDltAccountIds,
		SmartContractId:      cmd.Dispatch.SmartContractId,
		SmartContractName:    cmd.Dispatch.SmartContractName,
		MethodName:           cmd.Dispatch.MethodName,
		MethodArgs:           methodArgs,
		NetworkId:            cmd.NetworkId,
		ResolveNestedCalls:   true,
	})
	if err != nil {
		return nil, err
	}

	resp := result.(*signandsend.Response)

	return &signandsendbatchcross.CrossResponse{TransactionSentCrossEvent: mapper.ToCrossEvent(resp)}, nil
}

func requireCalls(calls []signandsendbatchcross.BatchCall) error {
	if len(calls) == 0 {
		return domainerrors.NewNestedCallsRequiredDomainError()
	}
	return nil
}

func dispatchArgs(dispatch *signandsendbatchcross.DispatchCall, calls []signandsendbatchcross.BatchCall) (map[string]any, error) {
	if dispatch.CallDataArgName == "" {
		return nil, domainerrors.NewCallDataArgNameRequiredDomainError()
	}

	methodArgs := make(map[string]any, len(dispatch.MethodArgs)+1)
	for name, value := range dispatch.MethodArgs {
		methodArgs[name] = value
	}
	if _, taken := methodArgs[dispatch.CallDataArgName]; taken {
		return nil, domainerrors.NewCallDataArgNameTakenDomainError(dispatch.CallDataArgName)
	}

	invocations := make([]portcommon.Invocation, 0, len(calls))
	for _, call := range calls {
		invocations = append(invocations, portcommon.Invocation{
			SmartContractName: call.SmartContractName,
			MethodName:        call.MethodName,
			MethodArgs:        call.MethodArgs,
		})
	}
	methodArgs[dispatch.CallDataArgName] = invocations

	return methodArgs, nil
}
