package createkeycross

import (
	"context"
	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/app/command/createkey"
	"dlt-ingress/src/main/dltingress/port/event/custodykey"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type CrossCommandAdapter struct {
	commandBus command.Bus
}

func NewCrossCommandAdapter(commandBus command.Bus) *CrossCommandAdapter {
	return &CrossCommandAdapter{commandBus: commandBus}
}

func (h *CrossCommandAdapter) Execute(ctx context.Context, cmd *CrossCommand) (*CrossResponse, error) {
	result, err := h.commandBus.Dispatch(ctx, &createkey.Command{
		Dlt:             cmd.Dlt,
		CustodyProvider: config.AppConfig.DltIngress.Custody.Provider,
	})
	if err != nil {
		return nil, err
	}

	resp := result.(*createkey.Response)

	return &CrossResponse{
		KeyCreatedCrossEvent: custodykeyevents.KeyCreatedCrossEvent{
			BaseEvent:       *event.NewBaseEvent(),
			DltAccountId:    resp.DltAccountId,
			ExternalId:      resp.ExternalId,
			Dlt:             resp.Dlt,
			CustodyProvider: resp.CustodyProvider,
		},
	}, nil
}
