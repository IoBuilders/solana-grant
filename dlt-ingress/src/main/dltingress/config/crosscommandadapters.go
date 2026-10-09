package dltingressconfig

import (
	"dlt-ingress/src/main/dltingress/internal/infra/crosscommand/buildtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/crosscommand/createkey"
	"dlt-ingress/src/main/dltingress/internal/infra/crosscommand/signandsend"
	"dlt-ingress/src/main/dltingress/internal/infra/crosscommand/signandsendbatch"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
)

func InitCommandAdapters(commandBus command.Bus) []interface{} {
	return []interface{}{
		createkeyadapter.NewCrossCommandAdapter(commandBus),
		signandsendadapter.NewCrossCommandAdapter(commandBus),
		signandsendbatchadapter.NewCrossCommandAdapter(commandBus),
		buildtransactionadapter.NewCrossCommandAdapter(commandBus),
	}
}
