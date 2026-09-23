package adapters

import (
	"dlt-ingress/src/main/dltingress/port/command/buildtransaction"
	"dlt-ingress/src/main/dltingress/port/command/createkey"
	"dlt-ingress/src/main/dltingress/port/command/signandsend"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
)

func InitCommandAdapters(commandBus command.Bus) []interface{} {
	return []interface{}{
		createkeycross.NewCrossCommandAdapter(commandBus),
		signandsendcross.NewCrossCommandAdapter(commandBus),
		buildtransactioncross.NewCrossCommandAdapter(commandBus),
	}
}
