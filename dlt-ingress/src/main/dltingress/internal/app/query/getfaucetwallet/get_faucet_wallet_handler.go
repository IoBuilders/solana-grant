package getfaucetwallet

import (
	"context"

	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	"dlt-ingress/src/main/dltingress/internal/domain/faucetwallet"
	"dlt-ingress/src/main/dltingress/internal/infra/evm"

	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/infra/svm"

	"dlt-ingress/src/main/config"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
)

type QueryHandler struct {
	faucetWalletRepo  faucetwallet.Repository
	custodyKeyRepo    custodykey.Repository
	evmClientRegistry evm.ClientRegistry
	svmClientRegistry svm.ClientRegistry
}

func NewHandler(
	faucetWalletRepo faucetwallet.Repository,
	custodyKeyRepo custodykey.Repository,
	evmClientRegistry evm.ClientRegistry,
	svmClientRegistry svm.ClientRegistry,
) *QueryHandler {
	return &QueryHandler{
		faucetWalletRepo:  faucetWalletRepo,
		custodyKeyRepo:    custodyKeyRepo,
		evmClientRegistry: evmClientRegistry,
		svmClientRegistry: svmClientRegistry,
	}
}

func (h *QueryHandler) Execute(ctx context.Context, query Query) (query.Response, error) {

	networkConfig, err := config.DltIngressConfig.DltIngress.GetDltIngressNetwork(query.NetworkId)
	if err != nil {
		return nil, err
	}

	dlt, err := common.ParseDlt(networkConfig.Dlt)
	if err != nil {
		return nil, err
	}

	wallet, err := h.faucetWalletRepo.FindByNetworkId(ctx, query.NetworkId)
	if err != nil {
		return nil, err
	}

	key, err := h.custodyKeyRepo.FindById(ctx, wallet.CustodyKeyId)
	if err != nil {
		return nil, err
	}

	balance, err := h.getBalance(ctx, dlt, query.NetworkId, key.DltAccountId)
	if err != nil {
		return nil, err
	}

	return Response{
		FaucetWalletId:   wallet.Id,
		NetworkId:        wallet.NetworkId,
		DltAccountId:     key.DltAccountId,
		FundingAmount:    wallet.FundingAmount,
		BalanceThreshold: wallet.BalanceThreshold,
		Balance:          *balance,
		Enabled:          wallet.Enabled,
	}, nil
}

func (h *QueryHandler) getBalance(ctx context.Context, dlt common.Dlt, networkId, dltAccountId string) (*amount.Amount, error) {
	switch dlt {
	case common.EVM:
		client, err := h.evmClientRegistry.GetClientForNetworkId(ctx, networkId)
		if err != nil {
			return nil, err
		}
		return client.GetBalance(ctx, dltAccountId)
	case common.SVM:
		client, err := h.svmClientRegistry.GetClientForNetworkId(networkId)
		if err != nil {
			return nil, err
		}
		return client.GetBalance(ctx, dltAccountId)
	default:
		return nil, domainerrors.NewInvalidDltDomainError(string(dlt))
	}
}
