package createkey

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	"dlt-ingress/src/main/dltingress/internal/infra/custody"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type CommandHandler struct {
	repo            custodykey.Repository
	custodyProvider custody.Port
	eventBus        event.Bus
}

func NewCommandHandler(
	repo custodykey.Repository,
	custodyProvider custody.Port,
	eventBus event.Bus,
) *CommandHandler {
	return &CommandHandler{
		repo:            repo,
		custodyProvider: custodyProvider,
		eventBus:        eventBus,
	}
}

func (h *CommandHandler) Handle(ctx context.Context, cmd *Command) (*Response, error) {
	dlt, err := common.ParseDlt(cmd.Dlt)
	if err != nil {
		return nil, err
	}
	if _, err = custodykey.ParseCustodyProvider(cmd.CustodyProvider); err != nil {
		return nil, err
	}

	keyType := dlt.KeyType()

	keyResponse, err := h.custodyProvider.CreateKey(ctx, &custody.CreateKeyRequest{
		KeyType: string(keyType),
		Dlt:     cmd.Dlt,
	})
	if err != nil {
		return nil, err
	}

	key, err := custodykey.NewCustodyKey(
		string(keyType),
		keyResponse.DltAccountId,
		cmd.Dlt,
		keyResponse.ExternalId,
		cmd.CustodyProvider,
	)
	if err != nil {
		return nil, err
	}

	if err := h.repo.Save(ctx, key); err != nil {
		return nil, err
	}

	evt := custodykey.KeyCreatedEvent{
		BaseEvent:       *event.NewBaseEvent(),
		DltAccountId:    key.DltAccountId,
		ExternalId:      key.ExternalId,
		Dlt:             string(key.Dlt),
		CustodyProvider: string(key.CustodyProvider),
	}

	if err := h.eventBus.Publish(ctx, evt); err != nil {
		return nil, err
	}

	return &Response{evt}, nil
}
