//go:build test

package trigger

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/broadcast"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/logging"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster/target"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// --- stubs ---

type stubEvent struct{}

func (stubEvent) EventType() event.Type { return event.TypeBlock }
func (stubEvent) NodeID() uuid.UUID     { return uuid.Nil }

// recordingLogger counts Error calls.
type recordingLogger struct {
	mu     sync.Mutex
	errors int
}

func (l *recordingLogger) Debug(string, ...any)                         {}
func (l *recordingLogger) Info(string, ...any)                          {}
func (l *recordingLogger) Warn(string, ...any)                          {}
func (l *recordingLogger) Error(string, ...any)                         {}
func (l *recordingLogger) DebugWithCtx(context.Context, string, ...any) {}
func (l *recordingLogger) InfoWithCtx(context.Context, string, ...any)  {}
func (l *recordingLogger) WarnWithCtx(context.Context, string, ...any)  {}
func (l *recordingLogger) ErrorWithCtx(context.Context, string, ...any) {
	l.mu.Lock()
	l.errors++
	l.mu.Unlock()
}

func (l *recordingLogger) errorCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.errors
}

// fakeRouter is a hand-written test double for routing.Router.
type fakeRouter struct {
	matched []*broadcaster.Broadcaster
	err     error
}

func (r *fakeRouter) Match(context.Context, event.Event) ([]*broadcaster.Broadcaster, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.matched, nil
}

func (r *fakeRouter) Reload(context.Context) error { return nil }

// fakeProducer is a hand-written test double for broadcast.Producer.
type fakeProducer struct {
	supportedType broadcaster.Type
	err           error
	produced      int
}

func (p *fakeProducer) Supports(t broadcaster.Type) bool { return t == p.supportedType }

func (p *fakeProducer) Produce(context.Context, broadcaster.Broadcaster, broadcaster.Configuration, event.Event) error {
	p.produced++
	return p.err
}

// fakeConfiguration is a hand-written test double for broadcaster.Configuration.
type fakeConfiguration struct {
	id  uuid.UUID
	typ broadcaster.Type
}

func (c fakeConfiguration) ID() uuid.UUID                                { return c.id }
func (c fakeConfiguration) Type() broadcaster.Type                       { return c.typ }
func (c fakeConfiguration) Validate() error                              { return nil }
func (c fakeConfiguration) AdditionalProperties() map[string]interface{} { return nil }

// fakeBroadcasterConfigurationConfigurationManager is a hand-written test
// double for configurationmanager.BroadcasterConfigurationConfigurationManager.
type fakeBroadcasterConfigurationConfigurationManager struct {
	configurations []broadcaster.Configuration
	err            error
}

func (m *fakeBroadcasterConfigurationConfigurationManager) Load(context.Context) ([]broadcaster.Configuration, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.configurations, nil
}

// RegisterMapper is never exercised by event_broadcaster_permanent_trigger, so the fake
// has nothing to record.
func (m *fakeBroadcasterConfigurationConfigurationManager) RegisterMapper(context.Context, string, func(ctx context.Context, source broadcaster.Configuration) (broadcaster.Configuration, error)) error {
	return nil
}

func mustBroadcaster(t *testing.T, configurationID uuid.UUID) *broadcaster.Broadcaster {
	t.Helper()
	destination, err := target.NewDestination("dest")
	require.NoError(t, err)
	tg, err := target.NewBlockTarget([]target.Destination{destination})
	require.NoError(t, err)
	b, err := broadcaster.NewBroadcaster(uuid.New(), tg, configurationID)
	require.NoError(t, err)
	return b
}

// --- tests ---

func TestEventBroadcasterPermanentTrigger_Process_ProducesToMatchedBroadcasters(t *testing.T) {
	configurationID := uuid.New()
	b := mustBroadcaster(t, configurationID)
	router := &fakeRouter{matched: []*broadcaster.Broadcaster{b}}
	manager := &fakeBroadcasterConfigurationConfigurationManager{
		configurations: []broadcaster.Configuration{fakeConfiguration{id: configurationID, typ: broadcaster.TypeHTTP}},
	}
	producer := &fakeProducer{supportedType: broadcaster.TypeHTTP}
	trg := NewEventBroadcasterPermanentTrigger(router, []broadcast.Producer{producer}, manager)
	require.NoError(t, trg.Reload(context.Background()))

	err := trg.Process(context.Background(), stubEvent{})

	require.NoError(t, err)
	assert.Equal(t, 1, producer.produced)
}

func TestEventBroadcasterPermanentTrigger_Process_SkipsBroadcasterWithNoMatchingProducer(t *testing.T) {
	configurationID := uuid.New()
	b := mustBroadcaster(t, configurationID)
	router := &fakeRouter{matched: []*broadcaster.Broadcaster{b}}
	manager := &fakeBroadcasterConfigurationConfigurationManager{
		configurations: []broadcaster.Configuration{fakeConfiguration{id: configurationID, typ: broadcaster.TypeHTTP}},
	}
	producer := &fakeProducer{supportedType: broadcaster.Type("OTHER")}
	logger := &recordingLogger{}
	logging.SetDefaultLogger(logger)
	trg := NewEventBroadcasterPermanentTrigger(router, []broadcast.Producer{producer}, manager)
	require.NoError(t, trg.Reload(context.Background()))

	err := trg.Process(context.Background(), stubEvent{})

	require.NoError(t, err)
	assert.Equal(t, 0, producer.produced)
	assert.Equal(t, 0, logger.errorCount())
}

func TestEventBroadcasterPermanentTrigger_Process_MissingConfiguration_LogsAndContinues(t *testing.T) {
	missingConfigurationID := uuid.New()
	validConfigurationID := uuid.New()
	missing := mustBroadcaster(t, missingConfigurationID)
	valid := mustBroadcaster(t, validConfigurationID)
	router := &fakeRouter{matched: []*broadcaster.Broadcaster{missing, valid}}
	manager := &fakeBroadcasterConfigurationConfigurationManager{
		configurations: []broadcaster.Configuration{fakeConfiguration{id: validConfigurationID, typ: broadcaster.TypeHTTP}},
	}
	producer := &fakeProducer{supportedType: broadcaster.TypeHTTP}
	logger := &recordingLogger{}
	logging.SetDefaultLogger(logger)
	trg := NewEventBroadcasterPermanentTrigger(router, []broadcast.Producer{producer}, manager)
	require.NoError(t, trg.Reload(context.Background()))

	err := trg.Process(context.Background(), stubEvent{})

	require.NoError(t, err)
	assert.Equal(t, 1, producer.produced, "the broadcaster with a valid configuration must still be produced to")
	assert.Equal(t, 1, logger.errorCount())
}

func TestEventBroadcasterPermanentTrigger_Process_ProduceError_LogsAndContinues(t *testing.T) {
	firstConfigurationID := uuid.New()
	secondConfigurationID := uuid.New()
	first := mustBroadcaster(t, firstConfigurationID)
	second := mustBroadcaster(t, secondConfigurationID)
	router := &fakeRouter{matched: []*broadcaster.Broadcaster{first, second}}
	manager := &fakeBroadcasterConfigurationConfigurationManager{
		configurations: []broadcaster.Configuration{
			fakeConfiguration{id: firstConfigurationID, typ: broadcaster.TypeHTTP},
			fakeConfiguration{id: secondConfigurationID, typ: broadcaster.TypeHTTP},
		},
	}
	producer := &fakeProducer{supportedType: broadcaster.TypeHTTP, err: errors.New("produce failed")}
	logger := &recordingLogger{}
	logging.SetDefaultLogger(logger)
	trg := NewEventBroadcasterPermanentTrigger(router, []broadcast.Producer{producer}, manager)
	require.NoError(t, trg.Reload(context.Background()))

	err := trg.Process(context.Background(), stubEvent{})

	require.NoError(t, err)
	assert.Equal(t, 2, producer.produced, "both broadcasters must be produced to despite the first failing")
	assert.Equal(t, 2, logger.errorCount())
}

func TestEventBroadcasterPermanentTrigger_Process_RouterError_LogsAndReturnsNil(t *testing.T) {
	router := &fakeRouter{err: errors.New("match failed")}
	manager := &fakeBroadcasterConfigurationConfigurationManager{}
	logger := &recordingLogger{}
	logging.SetDefaultLogger(logger)
	trg := NewEventBroadcasterPermanentTrigger(router, nil, manager)

	err := trg.Process(context.Background(), stubEvent{})

	require.NoError(t, err)
	assert.Equal(t, 1, logger.errorCount())
}

func TestEventBroadcasterPermanentTrigger_Supports_AlwaysTrue(t *testing.T) {
	trg := NewEventBroadcasterPermanentTrigger(&fakeRouter{}, nil, &fakeBroadcasterConfigurationConfigurationManager{})

	assert.True(t, trg.Supports(stubEvent{}))
}

func TestEventBroadcasterPermanentTrigger_Reload_PopulatesConfigurationSnapshot(t *testing.T) {
	configurationID := uuid.New()
	b := mustBroadcaster(t, configurationID)
	router := &fakeRouter{matched: []*broadcaster.Broadcaster{b}}
	manager := &fakeBroadcasterConfigurationConfigurationManager{}
	producer := &fakeProducer{supportedType: broadcaster.TypeHTTP}
	logger := &recordingLogger{}
	logging.SetDefaultLogger(logger)
	trg := NewEventBroadcasterPermanentTrigger(router, []broadcast.Producer{producer}, manager)

	require.NoError(t, trg.Process(context.Background(), stubEvent{}))
	assert.Equal(t, 0, producer.produced, "before Reload the configuration snapshot is empty")
	assert.Equal(t, 1, logger.errorCount())

	manager.configurations = []broadcaster.Configuration{fakeConfiguration{id: configurationID, typ: broadcaster.TypeHTTP}}
	require.NoError(t, trg.Reload(context.Background()))

	require.NoError(t, trg.Process(context.Background(), stubEvent{}))
	assert.Equal(t, 1, producer.produced, "after Reload the configuration must be found")
}

func TestEventBroadcasterPermanentTrigger_Reload_PropagatesManagerError(t *testing.T) {
	manager := &fakeBroadcasterConfigurationConfigurationManager{err: errors.New("load failed")}
	trg := NewEventBroadcasterPermanentTrigger(&fakeRouter{}, nil, manager)

	err := trg.Reload(context.Background())

	require.Error(t, err)
	assert.Empty(t, trg.snapshot())
}
