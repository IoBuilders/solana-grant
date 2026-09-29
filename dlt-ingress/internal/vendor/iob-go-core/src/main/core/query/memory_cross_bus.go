package query

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

type CrossBusInMemory struct {
	handlers map[string]any
	mu       sync.RWMutex
}

func NewCrossQueryBus() *CrossBusInMemory {
	return &CrossBusInMemory{
		handlers: make(map[string]any),
	}
}

func (qb *CrossBusInMemory) Register(handler any) error {
	logger.Info("Trying to register handler for cross query", "handler", fmt.Sprintf("%+v", handler))

	qb.mu.Lock()
	defer qb.mu.Unlock()

	queryName, err := validateCrossHandler(handler)

	if err != nil {
		logger.Error("[CROSS-QUERY-BUS] ⚠️ Error on validate handler", "queryName", queryName, "handler", fmt.Sprintf("%+v", handler), "error", err.Error())
		return err
	}

	if qb.handlers[queryName] != nil {
		logger.Warn("[CROSS-QUERY-BUS] ⚠️ Cross handler already registered for query", "queryName", queryName, "handler", fmt.Sprintf("%+v", handler))
	}

	qb.handlers[queryName] = handler
	logger.Info("Registered handler for cross query", "queryName", queryName, "handler", fmt.Sprintf("%+v", handler))
	return nil
}

func validateCrossHandler(handler any) (string, error) {

	var queryName string

	// Use reflection to extract the query type from the handler
	handlerType := reflect.TypeOf(handler)
	if handler == nil || handlerType.Kind() != reflect.Pointer ||
		handlerType.Elem().Kind() != reflect.Struct ||
		handlerType.NumMethod() == 0 {
		return queryName, fmt.Errorf("handler must be a pointer to a struct with Execute method")
	}

	//validates if contains Execute method
	handleMethod, exists := handlerType.MethodByName("Execute")
	if !exists {
		return queryName, fmt.Errorf("handler must implement Execute method")
	}

	//validates if method handle has 2 parameter  (ctx,query)
	queryType := handleMethod.Type.In(2)
	if !queryType.Implements(reflect.TypeOf((*CrossQuery)(nil)).Elem()) {
		return "", fmt.Errorf("query parameter must implement query interface")
	}

	//try to resolve sample query
	var concreteType reflect.Type

	if queryType.Kind() == reflect.Pointer {
		concreteType = queryType.Elem()
	} else {
		concreteType = queryType
	}

	sampleQuery, ok := reflect.New(concreteType).Interface().(CrossQuery)
	if !ok {
		return queryName, fmt.Errorf("failed to create sample command")
	}

	queryName = sampleQuery.CrossQueryName()

	return queryName, nil

}

// Deprecated: Use query.CrossAsk instead to benefit from strong typing
// and avoid manual type assertions
func (qb *CrossBusInMemory) Dispatch(ctx context.Context, query CrossQuery) (CrossResponse, error) {
	queryName := query.CrossQueryName()
	handler, exists := qb.handlers[queryName]
	if !exists {
		return nil, fmt.Errorf("no handler registered for cross query: %s. error dispatching command", queryName)

	}

	handlerValue := reflect.ValueOf(handler)
	method := handlerValue.MethodByName("Execute")

	results := method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(query)})

	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	return results[0].Interface(), nil
}
