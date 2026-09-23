//go:build test

package node

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/broadcast"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/logging"
	appstore "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/store"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster/target"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/common"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	domainfilter "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/httpclient"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/node"
)

// --- stubs ---

// fakeCollectionConfigManager is a hand-written test double satisfying every
// configurationmanager.CollectionConfigurationManager[T]-shaped port
// (NodeConfigurationManager, BroadcasterConfigurationManager,
// BroadcasterConfigurationConfigurationManager, FilterConfigurationManager)
// structurally, without one bespoke fake per port.
type fakeCollectionConfigManager[T any] struct {
	mu     sync.Mutex
	calls  int
	result []T
	err    error
}

func (m *fakeCollectionConfigManager[T]) Load(context.Context) ([]T, error) {
	m.mu.Lock()
	m.calls++
	m.mu.Unlock()
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func (m *fakeCollectionConfigManager[T]) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

// RegisterMapper satisfies configurationmanager.BroadcasterConfigurationConfigurationManager
// when fakeCollectionConfigManager is instantiated as fakeCollectionConfigManager[broadcaster.Configuration].
// node_initializer never calls it, so there's nothing for the fake to record.
func (m *fakeCollectionConfigManager[T]) RegisterMapper(context.Context, string, func(ctx context.Context, source broadcaster.Configuration) (broadcaster.Configuration, error)) error {
	return nil
}

// fakeConfigManager is the single-value counterpart, satisfying
// configurationmanager.ConfigurationManager[T]-shaped ports
// (HttpClientConfigurationManager).
type fakeConfigManager[T any] struct {
	mu     sync.Mutex
	calls  int
	result T
	err    error
}

func (m *fakeConfigManager[T]) Load(context.Context) (T, error) {
	m.mu.Lock()
	m.calls++
	m.mu.Unlock()
	if m.err != nil {
		var zero T
		return zero, m.err
	}
	return m.result, nil
}

func (m *fakeConfigManager[T]) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

// recordingLogger is a hand-written test double for logging.Logger that
// counts Error/ErrorWithCtx calls, letting tests assert every failing node
// in a batch was logged rather than the loop bailing out after the first.
type recordingLogger struct {
	mu     sync.Mutex
	errors int
}

func (l *recordingLogger) Debug(string, ...any) {}
func (l *recordingLogger) Info(string, ...any)  {}
func (l *recordingLogger) Warn(string, ...any)  {}
func (l *recordingLogger) Error(string, ...any) {
	l.mu.Lock()
	l.errors++
	l.mu.Unlock()
}
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

// useRecordingLogger swaps the package-level default logger for a
// recordingLogger for the duration of the test.
func useRecordingLogger(t *testing.T) *recordingLogger {
	t.Helper()
	logger := &recordingLogger{}
	logging.SetDefaultLogger(logger)
	t.Cleanup(func() { logging.SetDefaultLogger(logging.NewSlogLogger(slog.Default())) })
	return logger
}

// fakeProducer is a hand-written test double for broadcast.Producer that
// records how many times Produce was called.
type fakeProducer struct {
	mu       sync.Mutex
	supports broadcaster.Type
	produced int
}

func (p *fakeProducer) Supports(t broadcaster.Type) bool { return t == p.supports }

func (p *fakeProducer) Produce(context.Context, broadcaster.Broadcaster, broadcaster.Configuration, event.Event) error {
	p.mu.Lock()
	p.produced++
	p.mu.Unlock()
	return nil
}

func (p *fakeProducer) producedCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.produced
}

// fakeBlockEventStore, fakeTransactionEventStore and fakeContractEventStore
// are hand-written, do-nothing test doubles for the store.BlockEventStore /
// TransactionEventStore / ContractEventStore ports, letting a eventStores map be
// populated without a real persistence adapter.
type fakeBlockEventStore[D event.BlockEvent] struct{}

func (fakeBlockEventStore[D]) Save(context.Context, D) error          { return nil }
func (fakeBlockEventStore[D]) Supports(store.Type, reflect.Type) bool { return true }
func (fakeBlockEventStore[D]) GetLatest(context.Context, uuid.UUID) (uint64, error) {
	return 0, nil
}

type fakeLatestBlockStore[D event.BlockEvent] struct{}

func (fakeLatestBlockStore[D]) Save(context.Context, D) error          { return nil }
func (fakeLatestBlockStore[D]) Supports(store.Type, reflect.Type) bool { return true }
func (fakeLatestBlockStore[D]) Get(context.Context, uuid.UUID) (uint64, error) {
	return 0, nil
}

var _ appstore.BlockEventStore[event.SolanaBlockEvent] = fakeBlockEventStore[event.SolanaBlockEvent]{}

type fakeTransactionEventStore[D event.TransactionEvent] struct{}

func (fakeTransactionEventStore[D]) Save(context.Context, D) error          { return nil }
func (fakeTransactionEventStore[D]) Supports(store.Type, reflect.Type) bool { return true }

var _ appstore.TransactionEventStore[event.SolanaTransactionEvent] = fakeTransactionEventStore[event.SolanaTransactionEvent]{}

type fakeContractEventStore[D event.ContractEvent] struct{}

func (fakeContractEventStore[D]) Save(context.Context, D) error          { return nil }
func (fakeContractEventStore[D]) Supports(store.Type, reflect.Type) bool { return true }

var _ appstore.ContractEventStore[event.SolanaContractEvent] = fakeContractEventStore[event.SolanaContractEvent]{}

// newSolanaEventStores builds the eventStores map initializeSolanaNode's POLL
// branch expects, keyed exactly the way persistencegormsetup.Setup keys the
// real one (concrete Solana event types only — never the BlockEvent
// interface) so a getStore lookup using the wrong type parameter fails the
// same way it would in production instead of being masked by an extra key.
func newSolanaEventStores() map[reflect.Type]any {
	return map[reflect.Type]any{
		reflect.TypeFor[event.SolanaBlockEvent]():       fakeBlockEventStore[event.SolanaBlockEvent]{},
		reflect.TypeFor[event.SolanaTransactionEvent](): fakeTransactionEventStore[event.SolanaTransactionEvent]{},
		reflect.TypeFor[event.SolanaContractEvent]():    fakeContractEventStore[event.SolanaContractEvent]{},
	}
}

func newSolanaLatestBlockStores() map[reflect.Type]any {
	return map[reflect.Type]any{
		reflect.TypeFor[event.SolanaBlockEvent](): fakeLatestBlockStore[event.SolanaBlockEvent]{},
	}
}

// newTestNodeInitializer wires a DefaultNodeInitializer for tests that only
// exercise Initialize's own orchestration (config loading, per-node
// dispatch to initializeNode). The unused ports (producers, decoder,
// retryer, eventStores) are never reached by those paths, since every case here
// fails before addSolanaNodeSpecificTriggers/getStore would touch them.
func newTestNodeInitializer(
	nodeConfigManager configurationmanager.NodeConfigurationManager,
	httpClientConfigManager configurationmanager.HttpClientConfigurationManager,
) *DefaultNodeInitializer {
	return NewDefaultNodeInitializer(
		nodeConfigManager,
		&fakeCollectionConfigManager[*broadcaster.Broadcaster]{},
		&fakeCollectionConfigManager[broadcaster.Configuration]{},
		&fakeCollectionConfigManager[domainfilter.Filter]{},
		httpClientConfigManager,
		&fakeCollectionConfigManager[store.Configuration]{},
		nil,
		nil,
		nil,
		nil,
		map[reflect.Type]any{},
		map[reflect.Type]any{},
	)
}

// newPubSubSolanaNode builds a minimal Solana node configured for the PUBSUB
// subscription method, which initializeSolanaNode does not (yet) support.
// Connection is deliberately left unset: the PubSub branch fails before
// n.Connection is ever read.
func newPubSubSolanaNode(t *testing.T) *node.Node {
	t.Helper()
	subscriptionConfig, err := node.NewBlockSubscriptionConfiguration(node.NewPubSubBlockSubscriptionMethodConfiguration(), 0)
	require.NoError(t, err)
	return &node.Node{ID: uuid.New(), Type: node.TypeSolana, Subscription: subscriptionConfig}
}

// newPollSolanaNode builds a fully valid Solana node configured for the
// POLL subscription method over an HTTP connection, i.e. the one shape
// initializeSolanaNode actually knows how to wire up end to end.
func newPollSolanaNode(t *testing.T) *node.Node {
	t.Helper()
	name, err := node.NewName("solana-mainnet")
	require.NoError(t, err)
	endpoint, err := common.NewConnectionEndpointFromURL("https://api.mainnet-beta.solana.com")
	require.NoError(t, err)
	connection, err := common.NewHttpConnection(endpoint, common.DefaultRetryConfiguration())
	require.NoError(t, err)
	method, err := node.NewPollBlockSubscriptionMethodConfiguration(time.Second)
	require.NoError(t, err)
	subscriptionConfig, err := node.NewBlockSubscriptionConfiguration(method, 0)
	require.NoError(t, err)
	n, err := node.NewSolanaNode(uuid.New(), name, connection, subscriptionConfig)
	require.NoError(t, err)
	return n
}

// --- tests ---

func TestDefaultNodeInitializer_Initialize_NoNodes_ReturnsEmptyRunners(t *testing.T) {
	nodeConfigManager := &fakeCollectionConfigManager[*node.Node]{}
	httpClientConfigManager := &fakeConfigManager[*httpclient.HttpClient]{result: &httpclient.HttpClient{}}
	i := newTestNodeInitializer(nodeConfigManager, httpClientConfigManager)

	runners, err := i.Initialize(context.Background())

	require.NoError(t, err)
	assert.Empty(t, runners)
	assert.Equal(t, 1, nodeConfigManager.callCount())
	assert.Equal(t, 1, httpClientConfigManager.callCount())
}

func TestDefaultNodeInitializer_Initialize_NodeConfigLoadError_PropagatesAndSkipsHttpClientLoad(t *testing.T) {
	loadErr := errors.New("boom")
	nodeConfigManager := &fakeCollectionConfigManager[*node.Node]{err: loadErr}
	httpClientConfigManager := &fakeConfigManager[*httpclient.HttpClient]{result: &httpclient.HttpClient{}}
	i := newTestNodeInitializer(nodeConfigManager, httpClientConfigManager)

	runners, err := i.Initialize(context.Background())

	assert.ErrorIs(t, err, loadErr)
	assert.Nil(t, runners)
	assert.Equal(t, 1, nodeConfigManager.callCount())
	assert.Equal(t, 0, httpClientConfigManager.callCount())
}

func TestDefaultNodeInitializer_Initialize_HttpClientLoadError_Propagates(t *testing.T) {
	loadErr := errors.New("boom")
	nodeConfigManager := &fakeCollectionConfigManager[*node.Node]{}
	httpClientConfigManager := &fakeConfigManager[*httpclient.HttpClient]{err: loadErr}
	i := newTestNodeInitializer(nodeConfigManager, httpClientConfigManager)

	runners, err := i.Initialize(context.Background())

	assert.ErrorIs(t, err, loadErr)
	assert.Nil(t, runners)
	assert.Equal(t, 1, nodeConfigManager.callCount())
	assert.Equal(t, 1, httpClientConfigManager.callCount())
}

func TestDefaultNodeInitializer_Initialize_UnsupportedNodeType_LogsAndSkipsNode(t *testing.T) {
	logger := useRecordingLogger(t)

	n := &node.Node{ID: uuid.New(), Type: node.Type("EVM")}
	nodeConfigManager := &fakeCollectionConfigManager[*node.Node]{result: []*node.Node{n}}
	httpClientConfigManager := &fakeConfigManager[*httpclient.HttpClient]{result: &httpclient.HttpClient{}}
	i := newTestNodeInitializer(nodeConfigManager, httpClientConfigManager)

	runners, err := i.Initialize(context.Background())

	require.NoError(t, err)
	assert.Empty(t, runners)
	assert.Equal(t, 1, logger.errorCount())
	assert.Equal(t, 1, httpClientConfigManager.callCount())
}

func TestDefaultNodeInitializer_Initialize_UnsupportedSubscriptionMethod_LogsAndSkipsNode(t *testing.T) {
	logger := useRecordingLogger(t)

	n := newPubSubSolanaNode(t)
	nodeConfigManager := &fakeCollectionConfigManager[*node.Node]{result: []*node.Node{n}}
	httpClientConfigManager := &fakeConfigManager[*httpclient.HttpClient]{result: &httpclient.HttpClient{}}
	i := newTestNodeInitializer(nodeConfigManager, httpClientConfigManager)

	runners, err := i.Initialize(context.Background())

	require.NoError(t, err)
	assert.Empty(t, runners)
	assert.Equal(t, 1, logger.errorCount())
	assert.Equal(t, 1, httpClientConfigManager.callCount())
}

func TestDefaultNodeInitializer_Initialize_MultipleFailingNodes_ProcessesAllAndSkipsEach(t *testing.T) {
	logger := useRecordingLogger(t)

	unsupportedType := &node.Node{ID: uuid.New(), Type: node.Type("EVM")}
	unsupportedMethod := newPubSubSolanaNode(t)
	nodeConfigManager := &fakeCollectionConfigManager[*node.Node]{result: []*node.Node{unsupportedType, unsupportedMethod}}
	httpClientConfigManager := &fakeConfigManager[*httpclient.HttpClient]{result: &httpclient.HttpClient{}}
	i := newTestNodeInitializer(nodeConfigManager, httpClientConfigManager)

	runners, err := i.Initialize(context.Background())

	require.NoError(t, err)
	assert.Empty(t, runners)
	assert.Equal(t, 2, logger.errorCount())
	assert.Equal(t, 1, httpClientConfigManager.callCount())
}

func TestDefaultNodeInitializer_Initialize_SolanaPollNode_ReturnsSolanaNodeRunner(t *testing.T) {
	n := newPollSolanaNode(t)
	nodeConfigManager := &fakeCollectionConfigManager[*node.Node]{result: []*node.Node{n}}
	httpClient := httpclient.NewHttpClient(10, time.Minute, time.Second, time.Second, time.Second, time.Second, time.Minute, false)
	httpClientConfigManager := &fakeConfigManager[*httpclient.HttpClient]{result: httpClient}

	i := NewDefaultNodeInitializer(
		nodeConfigManager,
		&fakeCollectionConfigManager[*broadcaster.Broadcaster]{},
		&fakeCollectionConfigManager[broadcaster.Configuration]{},
		&fakeCollectionConfigManager[domainfilter.Filter]{},
		httpClientConfigManager,
		&fakeCollectionConfigManager[store.Configuration]{},
		nil,
		nil,
		nil,
		nil,
		newSolanaEventStores(),
		newSolanaLatestBlockStores(),
	)

	runners, err := i.Initialize(context.Background())

	require.NoError(t, err)
	require.Len(t, runners, 1)
	solanaRunner, ok := runners[0].(*DefaultNodeRunner)
	require.True(t, ok)
	assert.Equal(t, n.ID, solanaRunner.GetNode().ID)
	assert.False(t, solanaRunner.IsRunning())
	assert.Equal(t, 1, nodeConfigManager.callCount())
	assert.Equal(t, 1, httpClientConfigManager.callCount())
}

// Both EventRoutingService and EventBroadcasterPermanentTrigger only ever match/produce against
// an in-memory snapshot that an explicit Reload populates — Initialize must call that Reload (via
// addSharedTriggers), or a configured broadcaster+configuration would never actually get used.
// Every fake config manager elsewhere in this file is empty by default, which can't tell "Reload
// was never called" apart from "there's nothing configured"; this test configures real, non-empty
// broadcaster/configuration values and asserts the event actually reaches a Producer.
func TestDefaultNodeInitializer_Initialize_BroadcasterConfigurationLoaded_EventIsProduced(t *testing.T) {
	n := newPollSolanaNode(t)
	nodeConfigManager := &fakeCollectionConfigManager[*node.Node]{result: []*node.Node{n}}
	httpClient := httpclient.NewHttpClient(10, time.Minute, time.Second, time.Second, time.Second, time.Second, time.Minute, false)
	httpClientConfigManager := &fakeConfigManager[*httpclient.HttpClient]{result: httpClient}

	configurationID := uuid.New()
	configuration, err := broadcaster.NewGenericConfiguration(configurationID, broadcaster.TypeHTTP, nil)
	require.NoError(t, err)

	allTarget, err := target.NewAllTarget([]target.Destination{"/events"})
	require.NoError(t, err)
	b, err := broadcaster.NewBroadcaster(uuid.New(), allTarget, configurationID)
	require.NoError(t, err)

	producer := &fakeProducer{supports: broadcaster.TypeHTTP}

	i := NewDefaultNodeInitializer(
		nodeConfigManager,
		&fakeCollectionConfigManager[*broadcaster.Broadcaster]{result: []*broadcaster.Broadcaster{b}},
		&fakeCollectionConfigManager[broadcaster.Configuration]{result: []broadcaster.Configuration{configuration}},
		&fakeCollectionConfigManager[domainfilter.Filter]{},
		httpClientConfigManager,
		&fakeCollectionConfigManager[store.Configuration]{},
		nil,
		[]broadcast.Producer{producer},
		nil,
		nil,
		newSolanaEventStores(),
		newSolanaLatestBlockStores(),
	)

	runners, err := i.Initialize(context.Background())
	require.NoError(t, err)
	require.Len(t, runners, 1)
	solanaRunner, ok := runners[0].(*DefaultNodeRunner)
	require.True(t, ok)

	blockEvent, err := event.NewSolanaBlockEvent(n.ID, 1, "hash", nil, nil)
	require.NoError(t, err)
	solanaRunner.GetDispatcher().Dispatch(context.Background(), blockEvent)

	assert.Equal(t, 1, producer.producedCount())
}
