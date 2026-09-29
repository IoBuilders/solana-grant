package command

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"sync"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type CrossBusInMemory struct {
	ports                       map[string]any
	mu                          sync.RWMutex
	excludedCommandsFromInfoLog []string
	metricsRegistry             *metrics.Registry
}

func NewCrossCommandBus(excludedCommandsFromInfoLog []string, metricsRegistry *metrics.Registry) *CrossBusInMemory {
	return &CrossBusInMemory{
		ports:                       make(map[string]any),
		excludedCommandsFromInfoLog: excludedCommandsFromInfoLog,
		metricsRegistry:             metricsRegistry,
	}
}

func (cb *CrossBusInMemory) RegisterPort(port any) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	portType := reflect.TypeOf(port)
	if portType.Kind() != reflect.Pointer ||
		portType.Elem().Kind() != reflect.Struct ||
		portType.NumMethod() == 0 {
		return fmt.Errorf("port must be a pointer to a struct with Execute method")
	}

	portMethod, exists := portType.MethodByName("Execute")
	if !exists {
		return fmt.Errorf("port must implement Execute method")
	}

	// Get the command type from the Execute method's first parameter
	if portMethod.Type.NumIn() != 3 { // receiver + context + command
		return fmt.Errorf("execute method must have exactly (ctx, command)")
	}

	commandType := portMethod.Type.In(2)
	if !commandType.Implements(reflect.TypeOf((*CrossCommand)(nil)).Elem()) {
		return fmt.Errorf("command parameter must implement CrossCommand interface")
	}

	sampleCmd, ok := reflect.New(commandType.Elem()).Interface().(CrossCommand)
	if !ok {
		return fmt.Errorf("failed to create sample command")
	}

	commandName := sampleCmd.CrossCommandName()
	cb.ports[commandName] = port
	logger.Info("Registered port for command", "commandName", commandName, "port", port)
	return nil
}

func (cb *CrossBusInMemory) Dispatch(ctx context.Context, command CrossCommand) (CrossResponse, error) {
	commandName := command.CrossCommandName()
	port, exists := cb.ports[commandName]
	if !exists {
		return nil, fmt.Errorf("no port registered for command: %s", commandName)
	}

	logMessage := "[CROSS-COMMAND-BUS] Executing..."
	logArgs := []any{"commandName", commandName, "portName", reflect.TypeOf(port).String()}
	if slices.Contains(cb.excludedCommandsFromInfoLog, commandName) {
		logger.DebugWithCtx(ctx, logMessage, logArgs...)
	} else {
		logger.InfoWithCtx(ctx, logMessage, logArgs...)
	}

	portValue := reflect.ValueOf(port)
	method := portValue.MethodByName("Execute")
	if !method.IsValid() {
		return nil, fmt.Errorf("port doesn't implement Execute method")
	}

	start := time.Now()
	results := method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(command)})
	commandSuccess := isNilable(results[1]) && results[1].IsNil()
	duration := time.Since(start).Seconds()

	histogram, errHistogram := cb.metricsRegistry.GetHistogram(metrics.CommandMetricHistogram)

	if errHistogram != nil {
		logger.WarnWithCtx(ctx, "error on getting metrics histogram", "error", errHistogram, "commandName", commandName)
	} else if histogram != nil {
		histogram.Record(ctx, duration, metric.WithAttributes(
			attribute.Bool("cross", true),
			attribute.Bool("success", commandSuccess),
			attribute.String("command", commandName)))
	}

	if len(results) != 2 {
		return nil, fmt.Errorf("port returned unexpected number of results")
	}

	var response CrossResponse

	if commandSuccess {
		response = results[0].Interface().(CrossResponse)
	} else {
		err, _ := results[1].Interface().(error)
		logger.ErrorWithCtx(ctx, "[CROSS-COMMAND-BUS] error executing command", "portValue", portValue.String(), "commandName", commandName, "error", err)
		return response, err
	}

	return response, nil
}

func isNilable(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Interface, reflect.Slice, reflect.UnsafePointer:
		return true
	default:
		return false
	}
}
