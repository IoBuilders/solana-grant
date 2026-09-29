package query

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

type BusInMemory struct {
	handlers map[string]any
	mu       sync.RWMutex
}

func NewQueryBus() *BusInMemory {
	return &BusInMemory{
		handlers: make(map[string]any),
	}
}

func (qb *BusInMemory) Register(handler any) error {
	logger.Info("Trying to register handler for query", "handler", fmt.Sprintf("%+v", handler))

	qb.mu.Lock()
	defer qb.mu.Unlock()

	queryName, err := validateQueryHandler(handler)

	if err != nil {
		logger.Error("[QUERY-BUS] ⚠️ Error on validate handler", "queryName", queryName, "handler", fmt.Sprintf("%+v", handler), "error", err.Error())
		return err
	}

	if qb.handlers[queryName] != nil {
		logger.Warn("[QUERY-BUS] ⚠️ Query handler already registered for query", "queryName", queryName, "handler", fmt.Sprintf("%+v", handler))
	}

	qb.handlers[queryName] = handler
	logger.Info("Registered handler for query", "queryName", queryName, "handler", fmt.Sprintf("%+v", handler))

	return nil
}

func validateQueryHandler(handler any) (string, error) {

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
	if !queryType.Implements(reflect.TypeOf((*Query)(nil)).Elem()) {
		return "", fmt.Errorf("query parameter must implement query interface")
	}

	//try to resolve sample query
	var concreteType reflect.Type

	if queryType.Kind() == reflect.Pointer {
		concreteType = queryType.Elem()
	} else {
		concreteType = queryType
	}

	sampleQuery, ok := reflect.New(concreteType).Interface().(Query)
	if !ok {
		return queryName, fmt.Errorf("failed to create sample command")
	}

	queryName = sampleQuery.QueryName()

	return queryName, nil

}

// Deprecated: Use query.Ask instead to benefit from strong typing
// and avoid manual type assertions
func (qb *BusInMemory) Dispatch(ctx context.Context, query Query) (Response, error) {
	queryName := query.QueryName()
	handler, exists := qb.handlers[queryName]
	if !exists {
		return nil, fmt.Errorf("no handler registered for query: %s", queryName)
	}

	handlerValue := reflect.ValueOf(handler)
	method := handlerValue.MethodByName("Execute")

	results := method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(query)})

	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	return results[0].Interface(), nil
}
