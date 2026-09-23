package transanctionsender

import (
	"context"
	"dlt-ingress/src/main/dltingress/port/evm"
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

type EvmTransactionSender struct {
	evmClientRegistry evm.ClientRegistry
}

func NewEvmTransactionSender(evmClientRegistry evm.ClientRegistry) *EvmTransactionSender {
	return &EvmTransactionSender{
		evmClientRegistry,
	}
}

func (s *EvmTransactionSender) SendTransaction(ctx context.Context, request SendTransactionRequest) (*SendTransactionResponse, error) {
	evmClient, err := s.evmClientRegistry.GetClientForNetworkId(ctx, request.NetworkId)
	if err != nil {
		return nil, err
	}

	var hash string
	if err := evmClient.CallRpcMethod(ctx, &hash, "eth_sendRawTransaction", request.SignedTransaction); err != nil {
		logger.ErrorWithCtx(ctx, "error calling eth_sendRawTransaction", "error", err)
		return nil, fmt.Errorf("error sending transaction to network: %w", err)
	}

	return &SendTransactionResponse{TxId: hash}, nil
}

var _ Port = (*EvmTransactionSender)(nil)
