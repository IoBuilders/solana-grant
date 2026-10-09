package transanctionsender

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/infra/svm"
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

type SvmTransactionSender struct {
	svmClientRegistry svm.ClientRegistry
}

func NewSvmTransactionSender(svmClientRegistry svm.ClientRegistry) *SvmTransactionSender {
	return &SvmTransactionSender{svmClientRegistry}
}

func (s *SvmTransactionSender) SendTransaction(ctx context.Context, request SendTransactionRequest) (*SendTransactionResponse, error) {
	svmClient, err := s.svmClientRegistry.GetClientForNetworkId(request.NetworkId)
	if err != nil {
		return nil, err
	}

	txId, err := svmClient.SendTransaction(ctx, request.SignedTransaction)
	if err != nil {
		logger.ErrorWithCtx(ctx, "error calling sendTransaction", "error", err)
		return nil, fmt.Errorf("error sending transaction to network: %w", err)
	}

	return &SendTransactionResponse{TxId: txId}, nil
}

var _ Port = (*SvmTransactionSender)(nil)
