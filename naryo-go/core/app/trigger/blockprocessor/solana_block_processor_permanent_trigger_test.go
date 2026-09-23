//go:build test

package blockprocessor

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/logging"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/trigger"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/decoder"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter/solana"
)

// --- stubs ---

type stubEvent struct{}

func (stubEvent) EventType() event.Type { return event.TypeSlot }
func (stubEvent) NodeID() uuid.UUID     { return uuid.Nil }

// stubFilter is a minimal filter.Filter that is deliberately not a
// *filter.EventFilter, used to verify non-EVENT/non-EventFilter entries are
// ignored.
type stubFilter struct {
	nodeID uuid.UUID
}

func (f stubFilter) ID() uuid.UUID           { return uuid.New() }
func (f stubFilter) Name() filter.Name       { return "stub" }
func (f stubFilter) NodeID() uuid.UUID       { return f.nodeID }
func (f stubFilter) Type() filter.FilterType { return filter.FilterTypeTransaction }

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

// decodeCall records a single Decoder.Decode invocation.
type decodeCall struct {
	spec filter.Specification
	raw  decoder.RawPayload
}

// fakeDecoder is a hand-written test double for decoder.Decoder. A nil
// parameters slice (the zero value) reports "no match", matching the real
// Decoder contract.
type fakeDecoder struct {
	calls      []decodeCall
	parameters []parameter.ContractEventParameter
	err        error
}

func (d *fakeDecoder) Decode(spec filter.Specification, raw decoder.RawPayload) ([]parameter.ContractEventParameter, error) {
	d.calls = append(d.calls, decodeCall{spec: spec, raw: raw})
	if d.err != nil {
		return nil, d.err
	}
	return d.parameters, nil
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

// recordingLogger counts Error calls.
type recordingLogger struct {
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
	l.errors++
}

// --- helpers ---

func mustFilterName(t *testing.T, value string) filter.Name {
	t.Helper()
	name, err := filter.NewName(value)
	require.NoError(t, err)
	return name
}

func mustAnchorSpecification(t *testing.T, accountAddress string) *solana.AnchorSpecification {
	t.Helper()
	spec, err := solana.NewAnchorSpecification(solana.AnchorSourceInstruction, &accountAddress, "EventName", nil)
	require.NoError(t, err)
	return spec
}

func mustAnchorSpecificationAnyAccount(t *testing.T) *solana.AnchorSpecification {
	t.Helper()
	spec, err := solana.NewAnchorSpecification(solana.AnchorSourceInstruction, nil, "EventName", nil)
	require.NoError(t, err)
	return spec
}

func mustSplNativeSpecification(t *testing.T, mintAddress string) solana.SplNativeSpecification {
	t.Helper()
	spec, err := solana.NewSplNativeSpecification(solana.SplInstructionTransfer, &mintAddress)
	require.NoError(t, err)
	return spec
}

func mustEventFilter(t *testing.T, nodeID uuid.UUID, scope filter.Scope, contractAddress *string, spec filter.Specification) *filter.EventFilter {
	t.Helper()
	ef, err := filter.NewEventFilter(
		uuid.New(),
		mustFilterName(t, "test-filter"),
		nodeID,
		scope,
		contractAddress,
		spec,
		nil,
		nil,
	)
	require.NoError(t, err)
	return ef
}

func mustSolanaTransaction(t *testing.T, nodeID uuid.UUID, signature string, instructions []event.InstructionData) event.SolanaTransaction {
	t.Helper()
	tx, err := event.NewSolanaTransaction(nodeID, signature, 42, instructions, nil, nil, nil)
	require.NoError(t, err)
	return tx
}

func mustSolanaBlockEvent(t *testing.T, nodeID uuid.UUID, transactions []event.SolanaTransaction) event.SolanaBlockEvent {
	t.Helper()
	b, err := event.NewSolanaBlockEvent(nodeID, 42, "hash", nil, transactions)
	require.NoError(t, err)
	return b
}

func contractAddress(value string) *string { return &value }

// --- tests ---

func TestSolanaBlockProcessorPermanentTrigger_Process_MatchesDecodesAndDispatches(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransaction(t, nodeID, "sig", []event.InstructionData{
		{ProgramID: "prog-a", Data: []byte{1, 2, 3}, Accounts: []string{"acc-1"}},
	})
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	f := mustEventFilter(t, nodeID, filter.ScopeGlobal, nil, mustAnchorSpecification(t, "acc-1"))
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{f}}
	dec := &fakeDecoder{parameters: []parameter.ContractEventParameter{}}
	dispatcher := &fakeDispatcher{}
	helper := NewContractEventDispatcherHelper(dispatcher)
	logger := &recordingLogger{}
	logging.SetDefaultLogger(logger)

	trg := NewSolanaBlockProcessorPermanentTrigger(filterManager, dec, helper)

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	require.Len(t, dec.calls, 1)
	assert.Equal(t, decoder.RawPayload{ProgramID: "prog-a", Data: []byte{1, 2, 3}, Accounts: []string{"acc-1"}}, dec.calls[0].raw)

	require.Len(t, dispatcher.dispatched, 1)
	contractEvent, ok := dispatcher.dispatched[0].(event.SolanaContractEvent)
	require.True(t, ok)
	assert.Equal(t, nodeID, contractEvent.NodeID())
	assert.Equal(t, "prog-a", contractEvent.ProgramID)
	assert.Equal(t, "sig", contractEvent.Signature)
	assert.Equal(t, uint64(42), contractEvent.Slot)
	assert.Equal(t, "EventName", contractEvent.EventName())
	assert.Equal(t, event.ContractEventStatusConfirmed, contractEvent.Status())
	assert.Equal(t, 0, logger.errors)
}

func TestSolanaBlockProcessorPermanentTrigger_Process_ContractScopeFilter_OnlyMatchesItsOwnProgram(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransaction(t, nodeID, "sig", []event.InstructionData{
		{ProgramID: "prog-a", Accounts: []string{"acc-1"}},
		{ProgramID: "prog-b", Accounts: []string{"acc-1"}},
	})
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	f := mustEventFilter(t, nodeID, filter.ScopeContract, contractAddress("prog-a"), mustAnchorSpecification(t, "acc-1"))
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{f}}
	dec := &fakeDecoder{parameters: []parameter.ContractEventParameter{}}
	dispatcher := &fakeDispatcher{}
	helper := NewContractEventDispatcherHelper(dispatcher)

	trg := NewSolanaBlockProcessorPermanentTrigger(filterManager, dec, helper)

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	require.Len(t, dec.calls, 1, "decoder must only run for the matching program")
	require.Len(t, dispatcher.dispatched, 1)
	contractEvent := dispatcher.dispatched[0].(event.SolanaContractEvent)
	assert.Equal(t, "prog-a", contractEvent.ProgramID)
}

func TestSolanaBlockProcessorPermanentTrigger_Process_AnchorAccountAddressMismatch_DoesNotDecode(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransaction(t, nodeID, "sig", []event.InstructionData{
		{ProgramID: "prog-a", Accounts: []string{"acc-other"}},
	})
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	f := mustEventFilter(t, nodeID, filter.ScopeGlobal, nil, mustAnchorSpecification(t, "acc-1"))
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{f}}
	dec := &fakeDecoder{parameters: []parameter.ContractEventParameter{}}
	dispatcher := &fakeDispatcher{}

	trg := NewSolanaBlockProcessorPermanentTrigger(filterManager, dec, NewContractEventDispatcherHelper(dispatcher))

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	assert.Empty(t, dec.calls, "the instruction's accounts don't include the specification's AccountAddress")
	assert.Empty(t, dispatcher.dispatched)
}

func TestSolanaBlockProcessorPermanentTrigger_Process_AnchorNilAccountAddress_MatchesAnyAccount(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransaction(t, nodeID, "sig", []event.InstructionData{
		{ProgramID: "prog-a", Accounts: []string{"acc-other"}},
	})
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	f := mustEventFilter(t, nodeID, filter.ScopeGlobal, nil, mustAnchorSpecificationAnyAccount(t))
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{f}}
	dec := &fakeDecoder{parameters: []parameter.ContractEventParameter{}}
	dispatcher := &fakeDispatcher{}

	trg := NewSolanaBlockProcessorPermanentTrigger(filterManager, dec, NewContractEventDispatcherHelper(dispatcher))

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	assert.Len(t, dec.calls, 1, "a nil AccountAddress must match regardless of which accounts the instruction touches")
}

func TestSolanaBlockProcessorPermanentTrigger_Process_SplNativeMintAddressMatch_Decodes(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransaction(t, nodeID, "sig", []event.InstructionData{
		{ProgramID: "prog-a", Accounts: []string{"mint-1"}},
	})
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	f := mustEventFilter(t, nodeID, filter.ScopeGlobal, nil, mustSplNativeSpecification(t, "mint-1"))
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{f}}
	dec := &fakeDecoder{parameters: []parameter.ContractEventParameter{}}

	trg := NewSolanaBlockProcessorPermanentTrigger(filterManager, dec, NewContractEventDispatcherHelper(&fakeDispatcher{}))

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	assert.Len(t, dec.calls, 1)
}

func TestSolanaBlockProcessorPermanentTrigger_Process_SplNativeMintAddressMismatch_DoesNotDecode(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransaction(t, nodeID, "sig", []event.InstructionData{
		{ProgramID: "prog-a", Accounts: []string{"mint-other"}},
	})
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	f := mustEventFilter(t, nodeID, filter.ScopeGlobal, nil, mustSplNativeSpecification(t, "mint-1"))
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{f}}
	dec := &fakeDecoder{parameters: []parameter.ContractEventParameter{}}

	trg := NewSolanaBlockProcessorPermanentTrigger(filterManager, dec, NewContractEventDispatcherHelper(&fakeDispatcher{}))

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	assert.Empty(t, dec.calls, "the instruction's accounts don't include the specification's MintAddress")
}

func TestSolanaBlockProcessorPermanentTrigger_Process_NoMatch_DoesNotDispatch(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransaction(t, nodeID, "sig", []event.InstructionData{{ProgramID: "prog-a", Accounts: []string{"acc-1"}}})
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	f := mustEventFilter(t, nodeID, filter.ScopeGlobal, nil, mustAnchorSpecification(t, "acc-1"))
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{f}}
	dec := &fakeDecoder{parameters: nil} // nil == no match
	dispatcher := &fakeDispatcher{}
	helper := NewContractEventDispatcherHelper(dispatcher)

	trg := NewSolanaBlockProcessorPermanentTrigger(filterManager, dec, helper)

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	assert.Empty(t, dispatcher.dispatched)
}

func TestSolanaBlockProcessorPermanentTrigger_Process_DecodeError_LogsAndContinues(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransaction(t, nodeID, "sig", []event.InstructionData{
		{ProgramID: "prog-a", Accounts: []string{"acc-1"}},
		{ProgramID: "prog-a", Accounts: []string{"acc-1"}},
	})
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	f := mustEventFilter(t, nodeID, filter.ScopeGlobal, nil, mustAnchorSpecification(t, "acc-1"))
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{f}}
	dec := &fakeDecoder{err: errors.New("decode failed")}
	dispatcher := &fakeDispatcher{}
	helper := NewContractEventDispatcherHelper(dispatcher)
	logger := &recordingLogger{}
	logging.SetDefaultLogger(logger)

	trg := NewSolanaBlockProcessorPermanentTrigger(filterManager, dec, helper)

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err, "a per-instruction decode failure must not fail the whole block")
	assert.Empty(t, dispatcher.dispatched)
	assert.Len(t, dec.calls, 2, "both instructions must still be attempted")
	assert.Equal(t, 2, logger.errors)
}

func TestSolanaBlockProcessorPermanentTrigger_Process_SkipsEmptyBlock(t *testing.T) {
	nodeID := uuid.New()
	blockEvent := mustSolanaBlockEvent(t, nodeID, nil)

	filterManager := &fakeFilterConfigurationManager{}
	trg := NewSolanaBlockProcessorPermanentTrigger(filterManager, &fakeDecoder{}, NewContractEventDispatcherHelper(&fakeDispatcher{}))

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	assert.False(t, filterManager.loadCalled, "filters must not be loaded for an empty block")
}

func TestSolanaBlockProcessorPermanentTrigger_Process_NoActiveFilters_DoesNotCallDecoder(t *testing.T) {
	nodeID := uuid.New()
	tx := mustSolanaTransaction(t, nodeID, "sig", []event.InstructionData{{ProgramID: "prog-a", Accounts: []string{"acc-1"}}})
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	filterManager := &fakeFilterConfigurationManager{}
	dec := &fakeDecoder{}
	trg := NewSolanaBlockProcessorPermanentTrigger(filterManager, dec, NewContractEventDispatcherHelper(&fakeDispatcher{}))

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	assert.Empty(t, dec.calls)
}

func TestSolanaBlockProcessorPermanentTrigger_Process_IgnoresFiltersFromOtherNodesAndOtherTypes(t *testing.T) {
	nodeID := uuid.New()
	otherNodeID := uuid.New()
	tx := mustSolanaTransaction(t, nodeID, "sig", []event.InstructionData{{ProgramID: "prog-a", Accounts: []string{"acc-1"}}})
	blockEvent := mustSolanaBlockEvent(t, nodeID, []event.SolanaTransaction{tx})

	otherNodeFilter := mustEventFilter(t, otherNodeID, filter.ScopeGlobal, nil, mustAnchorSpecification(t, "acc-1"))
	filterManager := &fakeFilterConfigurationManager{filters: []filter.Filter{otherNodeFilter, stubFilter{nodeID: nodeID}}}
	dec := &fakeDecoder{parameters: []parameter.ContractEventParameter{}}

	trg := NewSolanaBlockProcessorPermanentTrigger(filterManager, dec, NewContractEventDispatcherHelper(&fakeDispatcher{}))

	err := trg.Process(context.Background(), blockEvent)

	require.NoError(t, err)
	assert.Empty(t, dec.calls)
}

func TestSolanaBlockProcessorPermanentTrigger_Process_IgnoresUnsupportedEvent(t *testing.T) {
	filterManager := &fakeFilterConfigurationManager{}
	trg := NewSolanaBlockProcessorPermanentTrigger(filterManager, &fakeDecoder{}, NewContractEventDispatcherHelper(&fakeDispatcher{}))

	err := trg.Process(context.Background(), stubEvent{})

	require.NoError(t, err)
	assert.False(t, filterManager.loadCalled)
}

func TestSolanaBlockProcessorPermanentTrigger_Supports(t *testing.T) {
	trg := NewSolanaBlockProcessorPermanentTrigger(&fakeFilterConfigurationManager{}, &fakeDecoder{}, NewContractEventDispatcherHelper(&fakeDispatcher{}))

	assert.True(t, trg.Supports(mustSolanaBlockEvent(t, uuid.New(), nil)))
	assert.False(t, trg.Supports(stubEvent{}))
}
