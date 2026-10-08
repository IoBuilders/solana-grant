package utils

import (
	"fmt"
	"reflect"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
)

func RegisterCommandHandlers(commandBus command.Bus, handlers []interface{}) {
	for _, h := range handlers {
		if err := commandBus.RegisterHandler(h); err != nil {
			panic(fmt.Errorf("failed to register handler %T: %w", h, err))
		}
	}
}

func RegisterQueryHandlers(queryBus query.Bus, handlers []any) {
	for _, handler := range handlers {
		if err := queryBus.Register(handler); err != nil {
			panic(fmt.Errorf("failed to register query handler %T: %w", handler, err))
		}
	}
}

func RegisterCrossCommandAdapters(crossCommandBus command.CrossBus, adapters []interface{}) {
	for _, h := range adapters {
		if err := crossCommandBus.RegisterPort(h); err != nil {
			panic(fmt.Errorf("failed to register port %T: %w", h, err))
		}
	}
}

func RegisterCrossQueryAdapters(crossQueryBus query.CrossBus, adapters []any) {
	for _, handler := range adapters {
		if err := crossQueryBus.Register(handler); err != nil {
			panic(fmt.Errorf("failed to register cross query handler %T: %w", handler, err))
		}
	}
}

func Key(c interface{}) string {
	return reflect.TypeOf(c).String()
}
