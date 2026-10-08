package buildtransaction

import (
	"context"
	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/app/service/txservice"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
)

type CommandHandler struct {
	txService *txservice.AppService
}

func NewCommandHandler(txService *txservice.AppService) *CommandHandler {
	return &CommandHandler{txService: txService}
}

func (h *CommandHandler) Handle(ctx context.Context, cmd *Command) (*Response, error) {
	network, err := config.DltIngressConfig.DltIngress.GetDltIngressNetwork(cmd.NetworkId)
	if err != nil {
		return nil, err
	}

	dlt, err := common.ParseDlt(network.Dlt)
	if err != nil {
		return nil, err
	}

	switch dlt {
	case common.EVM:
		return h.buildEvmUnsignedTransaction(ctx, cmd, *network)
	default:
		return nil, domainerrors.NewInvalidDltDomainError(network.Dlt)
	}
}

func (h *CommandHandler) buildEvmUnsignedTransaction(
	ctx context.Context,
	cmd *Command,
	network config.NetworkConfig,
) (*Response, error) {
	result, err := h.txService.PrepareTransaction(ctx, txservice.TransactionRequest{
		SenderDltAccountId:   cmd.SenderDltAccountId,
		SignersDltAccountIds: cmd.SignersDltAccountIds,
		SmartContractId:      cmd.SmartContractId,
		SmartContractName:    cmd.SmartContractName,
		MethodName:           cmd.MethodName,
		MethodArgs:           cmd.MethodArgs,
		NetworkId:            cmd.NetworkId,
	}, network)
	if err != nil {
		return nil, err
	}

	payload, err := portcommon.EncodeUnsignedTransaction(result.EVMTransactionResponse)
	if err != nil {
		return nil, err
	}

	return &Response{
		TransactionBuiltEvent: *transaction.NewTransactionBuiltEvent(cmd.SenderDltAccountId, payload),
	}, nil
}
