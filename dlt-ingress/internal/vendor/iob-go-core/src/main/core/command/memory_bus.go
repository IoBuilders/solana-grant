package command

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type BusInMemory struct {
	handlers                    map[string]any
	mu                          sync.RWMutex
	tm                          db.TransactionManager
	excludedCommandsFromInfoLog []string
	metricsRegistry             *metrics.Registry
}

func NewCommandBus(tm db.TransactionManager, excludedCommandsFromInfoLog []string, metricsRegistry *metrics.Registry) *BusInMemory {
	return &BusInMemory{
		handlers:                    make(map[string]any),
		tm:                          tm,
		excludedCommandsFromInfoLog: excludedCommandsFromInfoLog,
		metricsRegistry:             metricsRegistry,
	}
}

func (cb *BusInMemory) RegisterHandler(handler any) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() != reflect.Pointer ||
		handlerType.Elem().Kind() != reflect.Struct ||
		handlerType.NumMethod() == 0 {
		return fmt.Errorf("handler must be a pointer to a struct with Handle method")
	}

	handleMethod, exists := handlerType.MethodByName("Handle")
	if !exists {
		return fmt.Errorf("handler must implement Handle method")
	}

	if handleMethod.Type.NumIn() != 3 { // receiver + context + command
		return fmt.Errorf("handle method must have exactly one parameter")
	}

	commandType := handleMethod.Type.In(2)
	if !commandType.Implements(reflect.TypeFor[Command]()) {
		return fmt.Errorf("command parameter must implement Command interface")
	}

	// Create a sample command to validate and get its name
	sampleCmd, ok := reflect.New(commandType.Elem()).Interface().(Command)
	if !ok {
		return fmt.Errorf("failed to create sample command")
	}

	commandName := sampleCmd.CommandName()
	cb.handlers[commandName] = handler
	logger.Info("Registered handler for command", "commandName", commandName, "handler", handler)
	return nil
}

func (cb *BusInMemory) Dispatch(ctx context.Context, command Command) (Response, error) {
	commandName := command.CommandName()

	handler, exists := cb.handlers[commandName]
	if !exists {
		return nil, fmt.Errorf("no handler registered for command: %s", commandName)
	}

	logMessage := "[COMMAND-BUS] Executing..."
	logArgs := []any{"commandName", commandName, "handlerName", reflect.TypeOf(handler).String()}
	if slices.Contains(cb.excludedCommandsFromInfoLog, commandName) {
		logger.DebugWithCtx(ctx, logMessage, logArgs...)
	} else {
		logger.InfoWithCtx(ctx, logMessage, logArgs...)
	}

	handlerValue := reflect.ValueOf(handler)
	method := handlerValue.MethodByName("Handle")
	if !method.IsValid() {
		return nil, fmt.Errorf("handler doesn't implement Handle method")
	}

	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var tx db.Transaction
	// Do not open transactions for order commands as they do not modify state.
	// This is done to avoid database connection locks on order commands module databases due to ingress nonce lock when there is high concurrency.
	if !strings.HasPrefix(commandName, "order") {
		childCtx = cb.tm.WithNewTransactionValue(childCtx)
		tx = cb.tm.TransactionValue(childCtx)
	}

	start := time.Now()
	results := method.Call([]reflect.Value{reflect.ValueOf(childCtx), reflect.ValueOf(command)})
	commandSuccess := isNilable(results[1]) && results[1].IsNil()
	duration := time.Since(start).Seconds()

	histogram, errHistogram := cb.metricsRegistry.GetHistogram(metrics.CommandMetricHistogram)

	if errHistogram != nil {
		logger.WarnWithCtx(ctx, "error on getting metrics histogram", "error", errHistogram, "commandName", commandName)
	} else {
		histogram.Record(ctx, duration, metric.WithAttributes(
			attribute.Bool("cross", false),
			attribute.Bool("success", commandSuccess),
			attribute.String("command", commandName)))
	}

	if len(results) != 2 {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, fmt.Errorf("transaction rollback error: %w", rollbackErr)
		}
		return nil, fmt.Errorf("handler returned unexpected number of results")
	}

	var response Response

	if commandSuccess {
		response = results[0].Interface().(Response)
	} else {
		if tx != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return response, fmt.Errorf("transaction rollback error: %w", rollbackErr)
			}
		}
		err, _ := results[1].Interface().(error)
		logger.ErrorWithCtx(ctx, "[COMMAND-BUS] Error executing command", "handlerValue", handlerValue.String(), "commandName", commandName, "error", err)
		return response, err
	}

	if tx != nil {
		if commitErr := tx.Commit(); commitErr != nil {
			return response, fmt.Errorf("transaction commit error: %w", commitErr)
		}
	}

	return response, nil
}
