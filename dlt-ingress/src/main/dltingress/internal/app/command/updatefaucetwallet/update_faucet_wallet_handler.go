package updatefaucetwallet

import (
	"context"

	"dlt-ingress/src/main/dltingress/internal/domain/faucetwallet"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type CommandHandler struct {
	faucetWalletRepo faucetwallet.Repository
	eventBus         event.Bus
}

func NewCommandHandler(faucetWalletRepo faucetwallet.Repository, eventBus event.Bus) *CommandHandler {
	return &CommandHandler{faucetWalletRepo: faucetWalletRepo, eventBus: eventBus}
}

func (h *CommandHandler) Handle(ctx context.Context, cmd *Command) (*Response, error) {
	wallet, err := h.faucetWalletRepo.FindByNetworkId(ctx, cmd.NetworkId)
	if err != nil {
		return nil, err
	}

	if err := wallet.UpdateFundingAmount(cmd.FundingAmount); err != nil {
		return nil, err
	}

	if err := wallet.UpdateBalanceThreshold(cmd.BalanceThreshold); err != nil {
		return nil, err
	}

	wallet.SetEnabled(cmd.Enabled)

	if err := h.faucetWalletRepo.Save(ctx, wallet); err != nil {
		return nil, err
	}

	evt := faucetwallet.UpdatedEvent{
		BaseEvent:        *event.NewBaseEvent(),
		FaucetWalletId:   wallet.Id,
		NetworkId:        wallet.NetworkId,
		FundingAmount:    wallet.FundingAmount,
		BalanceThreshold: wallet.BalanceThreshold,
		Enabled:          wallet.Enabled,
	}

	if err := h.eventBus.Publish(ctx, evt); err != nil {
		return nil, err
	}

	return &Response{UpdatedEvent: evt}, nil
}
