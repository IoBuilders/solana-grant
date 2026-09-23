//go:build test

package transactionprocessor

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/trigger"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
)

// --- stubs ---

type stubEvent struct{}

func (stubEvent) EventType() event.Type { return event.TypeSlot }
func (stubEvent) NodeID() uuid.UUID     { return uuid.Nil }

// fakeFilterConfigurationManager is a hand-written test double for
// configurationmanager.FilterConfigurationManager.
type fakeFilterConfigurationManager struct {
	filters    []filter.Filter
	err        error
	loadCalled bool
}

func (m *fakeFilterConfigurationManager) Load(context.Context) ([]filter.Filter, error) {
	m.loadCalled = true
	if m.err != nil {
		return nil, m.err
	}
	return m.filters, nil
}

// fakeProgramErrorRegistryConfigurationManager is a hand-written test double
// for configurationmanager.ProgramErrorRegistryConfigurationManager.
type fakeProgramErrorRegistryConfigurationManager struct {
	registry event.ProgramErrorRegistry
	err      error
}

func (m *fakeProgramErrorRegistryConfigurationManager) Load(context.Context) (event.ProgramErrorRegistry, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.registry, nil
}

// fakeDispatcher is a hand-written test double for dispatch.Dispatcher.
type fakeDispatcher struct {
	dispatched []event.Event
}

func (d *fakeDispatcher) Dispatch(_ context.Context, e event.Event) {
	d.dispatched = append(d.dispatched, e)
}

func (d *fakeDispatcher) AddTrigger(trigger.Trigger) {}

func (d *fakeDispatcher) RemoveExpired() {}

// --- helpers ---

func mustFilterName(t *testing.T, value string) filter.Name {
	t.Helper()
	name, err := filter.NewName(value)
	require.NoError(t, err)
	return name
}

func mustTransactionFilter(t *testing.T, nodeID uuid.UUID, identifierType filter.IdentifierType, values ...string) *filter.TransactionFilter {
	t.Helper()
	tf, err := filter.NewTransactionFilter(uuid.New(), mustFilterName(t, "test-filter"), nodeID, identifierType, values, nil)
	require.NoError(t, err)
	return tf
}

func mustSolanaTransaction(t *testing.T, nodeID uuid.UUID, signature string) event.SolanaTransaction {
	t.Helper()
	tx, err := event.NewSolanaTransaction(nodeID, signature, 42, nil, nil, nil, nil)
	require.NoError(t, err)
	return tx
}

func mustSolanaTransactionWithAccounts(t *testing.T, nodeID uuid.UUID, signature string, accounts []string) event.SolanaTransaction {
	t.Helper()
	tx, err := event.NewSolanaTransaction(nodeID, signature, 42, nil, accounts, nil, nil)
	require.NoError(t, err)
	return tx
}

func mustSolanaTransactionWithErr(t *testing.T, nodeID uuid.UUID, signature string, errPayload string) event.SolanaTransaction {
	t.Helper()
	tx, err := event.NewSolanaTransaction(nodeID, signature, 42, nil, nil, nil, &errPayload)
	require.NoError(t, err)
	return tx
}

func mustSolanaBlockEvent(t *testing.T, nodeID uuid.UUID, transactions []event.SolanaTransaction) event.SolanaBlockEvent {
	t.Helper()
	b, err := event.NewSolanaBlockEvent(nodeID, 42, "hash", nil, transactions)
	require.NoError(t, err)
	return b
}

// --- tests ---

func TestSolanaTransactionProcessorPermanentTrigger_Process_DispatchesMatchingTransaction(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransaction(t, nodeID, "sig-123")
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	f := mustTransactionFilter(t, nodeID, filter.IdentifierTypeHash, "sig-123")
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{f}}
	dispatcher := &fakeDispatcher{}

	trg := NewSolanaTransactionProcessorPermanentTrigger(filterManager, dispatcher, &fakeProgramErrorRegistryConfigurationManager{})

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	require.Len(t, dispatcher.dispatched, 1)
	dispatched, ok := dispatcher.dispatched[0].(event.SolanaTransactionEvent)
	require.True(t, ok)
	assert.Equal(t, "sig-123", dispatched.Signature)
}

func TestSolanaTransactionProcessorPermanentTrigger_Process_DispatchesOncePerMatchingFilter(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransaction(t, nodeID, "sig-123")
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	f1 := mustTransactionFilter(t, nodeID, filter.IdentifierTypeHash, "sig-123")
	f2 := mustTransactionFilter(t, nodeID, filter.IdentifierTypeHash, "sig-123")
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{f1, f2}}
	dispatcher := &fakeDispatcher{}

	trg := NewSolanaTransactionProcessorPermanentTrigger(filterManager, dispatcher, &fakeProgramErrorRegistryConfigurationManager{})

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	assert.Len(t, dispatcher.dispatched, 2)
}

func TestSolanaTransactionProcessorPermanentTrigger_Process_NoMatch_DoesNotDispatch(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransaction(t, nodeID, "sig-123")
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	f := mustTransactionFilter(t, nodeID, filter.IdentifierTypeHash, "sig-other")
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{f}}
	dispatcher := &fakeDispatcher{}

	trg := NewSolanaTransactionProcessorPermanentTrigger(filterManager, dispatcher, &fakeProgramErrorRegistryConfigurationManager{})

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	assert.Empty(t, dispatcher.dispatched)
}

func TestSolanaTransactionProcessorPermanentTrigger_Process_SkipsEmptyBlock(t *testing.T) {
	nodeID := uuid.New()
	blockEvent := mustSolanaBlockEvent(t, nodeID, nil)

	filterManager := &fakeFilterConfigurationManager{}
	trg := NewSolanaTransactionProcessorPermanentTrigger(filterManager, &fakeDispatcher{}, &fakeProgramErrorRegistryConfigurationManager{})

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	assert.False(t, filterManager.loadCalled, "filters must not be loaded for an empty block")
}

func TestSolanaTransactionProcessorPermanentTrigger_Process_NoActiveFilters_DoesNotDispatch(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransaction(t, nodeID, "sig-123")
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	filterManager := &fakeFilterConfigurationManager{}
	dispatcher := &fakeDispatcher{}
	trg := NewSolanaTransactionProcessorPermanentTrigger(filterManager, dispatcher, &fakeProgramErrorRegistryConfigurationManager{})

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	assert.Empty(t, dispatcher.dispatched)
}

func TestSolanaTransactionProcessorPermanentTrigger_Process_IgnoresFiltersFromOtherNodesAndOtherTypes(t *testing.T) {
	nodeID := uuid.New()
	otherNodeID := uuid.New()
	tx := mustSolanaTransaction(t, nodeID, "sig-123")
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	otherNodeFilter := mustTransactionFilter(t, otherNodeID, filter.IdentifierTypeHash, "sig-123")
	eventFilter, err := filter.NewEventFilter(uuid.New(), mustFilterName(t, "event-filter"), nodeID, filter.ScopeGlobal, nil, stubSpecification{}, nil, nil)
	require.NoError(t, err)
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{otherNodeFilter, eventFilter}}
	dispatcher := &fakeDispatcher{}

	trg := NewSolanaTransactionProcessorPermanentTrigger(filterManager, dispatcher, &fakeProgramErrorRegistryConfigurationManager{})

	err = trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	assert.Empty(t, dispatcher.dispatched)
}

func TestSolanaTransactionProcessorPermanentTrigger_Process_IgnoresUnsupportedEvent(t *testing.T) {
	filterManager := &fakeFilterConfigurationManager{}
	trg := NewSolanaTransactionProcessorPermanentTrigger(filterManager, &fakeDispatcher{}, &fakeProgramErrorRegistryConfigurationManager{})

	err := trg.Process(context.Background(), stubEvent{})

	require.NoError(t, err)
	assert.False(t, filterManager.loadCalled)
}

func TestSolanaTransactionProcessorPermanentTrigger_Process_AddressesFilter_MatchesWhenAllAccountsPresent(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransactionWithAccounts(t, nodeID, "sig-123", []string{"acc-1", "acc-2"})
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	f := mustTransactionFilter(t, nodeID, filter.IdentifierTypeAddresses, "acc-1", "acc-2")
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{f}}
	dispatcher := &fakeDispatcher{}

	trg := NewSolanaTransactionProcessorPermanentTrigger(filterManager, dispatcher, &fakeProgramErrorRegistryConfigurationManager{})

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	assert.Len(t, dispatcher.dispatched, 1)
}

func TestSolanaTransactionProcessorPermanentTrigger_Process_AddressesFilter_NoMatchWhenOneAccountMissing(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransactionWithAccounts(t, nodeID, "sig-123", []string{"acc-1"})
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	f := mustTransactionFilter(t, nodeID, filter.IdentifierTypeAddresses, "acc-1", "acc-2")
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{f}}
	dispatcher := &fakeDispatcher{}

	trg := NewSolanaTransactionProcessorPermanentTrigger(filterManager, dispatcher, &fakeProgramErrorRegistryConfigurationManager{})

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	assert.Empty(t, dispatcher.dispatched)
}

func TestSolanaTransactionProcessorPermanentTrigger_Process_FailedTransaction_DecoratesWithDecodedReason(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransactionWithErr(t, nodeID, "sig-123", `"AccountInUse"`)
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	f := mustTransactionFilter(t, nodeID, filter.IdentifierTypeHash, "sig-123")
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{f}}
	dispatcher := &fakeDispatcher{}

	trg := NewSolanaTransactionProcessorPermanentTrigger(filterManager, dispatcher, &fakeProgramErrorRegistryConfigurationManager{})

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	require.Len(t, dispatcher.dispatched, 1)
	dispatched, ok := dispatcher.dispatched[0].(event.SolanaTransactionEvent)
	require.True(t, ok)
	require.NotNil(t, dispatched.DecodedReason)
	assert.Equal(t, "AccountInUse", *dispatched.DecodedReason)
}

func TestSolanaTransactionProcessorPermanentTrigger_Process_NonFailedTransaction_DispatchedWithoutDecodedReason(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransaction(t, nodeID, "sig-123")
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	f := mustTransactionFilter(t, nodeID, filter.IdentifierTypeHash, "sig-123")
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{f}}
	dispatcher := &fakeDispatcher{}

	trg := NewSolanaTransactionProcessorPermanentTrigger(filterManager, dispatcher, &fakeProgramErrorRegistryConfigurationManager{})

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	require.Len(t, dispatcher.dispatched, 1)
	dispatched, ok := dispatcher.dispatched[0].(event.SolanaTransactionEvent)
	require.True(t, ok)
	assert.Nil(t, dispatched.DecodedReason)
}

func TestSolanaTransactionProcessorPermanentTrigger_Process_RegistryLoadError_StillDispatchesUndecorated(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransactionWithErr(t, nodeID, "sig-123", `{"InstructionError":[0,{"Custom":6002}]}`)
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	f := mustTransactionFilter(t, nodeID, filter.IdentifierTypeHash, "sig-123")
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{f}}
	dispatcher := &fakeDispatcher{}
	registryManager := &fakeProgramErrorRegistryConfigurationManager{err: assert.AnError}

	trg := NewSolanaTransactionProcessorPermanentTrigger(filterManager, dispatcher, registryManager)

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	require.Len(t, dispatcher.dispatched, 1)
	dispatched, ok := dispatcher.dispatched[0].(event.SolanaTransactionEvent)
	require.True(t, ok)
	assert.Nil(t, dispatched.DecodedReason)
}

func TestSolanaTransactionProcessorPermanentTrigger_Supports(t *testing.T) {
	trg := NewSolanaTransactionProcessorPermanentTrigger(&fakeFilterConfigurationManager{}, &fakeDispatcher{}, &fakeProgramErrorRegistryConfigurationManager{})

	assert.True(t, trg.Supports(mustSolanaBlockEvent(t, uuid.New(), nil)))
	assert.False(t, trg.Supports(stubEvent{}))
}

// stubSpecification is a minimal filter.Specification, used only to build a
// *filter.EventFilter for the "ignores other filter types" test.
type stubSpecification struct{}

func (stubSpecification) Strategy() filter.Strategy        { return "STUB" }
func (stubSpecification) EventName() string                { return "stub-event" }
func (stubSpecification) Matches(event.ContractEvent) bool { return false }
