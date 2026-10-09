package createfaucetwallet

import (
	"context"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/domain/faucetwallet"
	"dlt-ingress/src/main/dltingress/internal/infra/custody"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type CommandHandler struct {
	faucetWalletRepo faucetwallet.Repository
	custodyKeyRepo   custodykey.Repository
	custodyProvider  custody.Port
	eventBus         event.Bus
}

func NewCommandHandler(
	faucetWalletRepo faucetwallet.Repository,
	custodyKeyRepo custodykey.Repository,
	custodyProvider custody.Port,
	eventBus event.Bus,
) *CommandHandler {
	return &CommandHandler{
		faucetWalletRepo: faucetWalletRepo,
		custodyKeyRepo:   custodyKeyRepo,
		custodyProvider:  custodyProvider,
		eventBus:         eventBus,
	}
}

func (h *CommandHandler) Handle(ctx context.Context, cmd *Command) (*Response, error) {
	networkConfig, err := config.DltIngressConfig.DltIngress.GetDltIngressNetwork(cmd.NetworkId)
	if err != nil {
		return nil, err
	}

	exists, err := h.faucetWalletRepo.ExistByNetworkId(ctx, cmd.NetworkId)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domainerrors.NewFaucetWalletAlreadyExistsDomainError(cmd.NetworkId)
	}

	dlt, err := common.ParseDlt(networkConfig.Dlt)
	if err != nil {
		return nil, err
	}

	keyResponse, err := h.custodyProvider.CreateKey(ctx, &custody.CreateKeyRequest{
		KeyType: string(dlt.KeyType()),
		Dlt:     networkConfig.Dlt,
	})
	if err != nil {
		return nil, err
	}

	key, err := custodykey.NewCustodyKey(
		string(dlt.KeyType()),
		keyResponse.DltAccountId,
		networkConfig.Dlt,
		keyResponse.ExternalId,
		config.DltIngressConfig.DltIngress.Custody.Provider,
	)
	if err != nil {
		return nil, err
	}

	if err := h.custodyKeyRepo.Save(ctx, key); err != nil {
		return nil, err
	}

	wallet, err := faucetwallet.NewFaucetWallet(cmd.NetworkId, key.Id, cmd.FundingAmount, cmd.BalanceThreshold)
	if err != nil {
		return nil, err
	}

	if err := h.faucetWalletRepo.Save(ctx, wallet); err != nil {
		return nil, err
	}

	evt := faucetwallet.CreatedEvent{
		BaseEvent:        *event.NewBaseEvent(),
		FaucetWalletId:   wallet.Id,
		NetworkId:        wallet.NetworkId,
		CustodyKeyId:     wallet.CustodyKeyId,
		DltAccountId:     key.DltAccountId,
		FundingAmount:    wallet.FundingAmount,
		BalanceThreshold: wallet.BalanceThreshold,
		Enabled:          wallet.Enabled,
	}

	if err := h.eventBus.Publish(ctx, evt); err != nil {
		return nil, err
	}

	return &Response{CreatedEvent: evt}, nil
}
