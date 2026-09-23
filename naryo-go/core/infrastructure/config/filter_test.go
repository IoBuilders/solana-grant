//go:build test

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	domainfilter "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter/solana"
)

const (
	validFilterID     = "00000000-0000-0000-0000-000000000001"
	validFilterNodeID = "00000000-0000-0000-0000-000000000002"
)

func ptr(s string) *string { return &s }

func TestFilterProperties_Map_InvalidID(t *testing.T) {
	fp := &FilterProperties{
		ID:     "not-a-uuid",
		NodeID: validFilterNodeID,
		Type:   domainfilter.FilterTypeTransaction.String(),
	}

	_, err := fp.Map()

	assert.Error(t, err)
}

func TestFilterProperties_Map_InvalidNodeID(t *testing.T) {
	fp := &FilterProperties{
		ID:     validFilterID,
		NodeID: "not-a-uuid",
		Type:   domainfilter.FilterTypeTransaction.String(),
	}

	_, err := fp.Map()

	assert.Error(t, err)
}

func TestFilterProperties_Map_UnsupportedType(t *testing.T) {
	fp := &FilterProperties{
		ID:     validFilterID,
		NodeID: validFilterNodeID,
		Type:   "UNKNOWN",
	}

	_, err := fp.Map()

	assert.ErrorContains(t, err, "unsupported filter type: UNKNOWN")
}

func TestFilterProperties_Map_TransactionFilter_Success(t *testing.T) {
	fp := &FilterProperties{
		ID:             validFilterID,
		NodeID:         validFilterNodeID,
		Name:           "tx-filter",
		Type:           domainfilter.FilterTypeTransaction.String(),
		IdentifierType: string(domainfilter.IdentifierTypeHash),
		Value:          []string{"0xabc123"},
		Statuses:       []string{domainfilter.TransactionStatusFailed.String()},
	}

	result, err := fp.Map()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, domainfilter.FilterTypeTransaction, result.Type())
	assert.Equal(t, domainfilter.Name("tx-filter"), result.Name())
	tf, ok := result.(*domainfilter.TransactionFilter)
	require.True(t, ok)
	assert.Equal(t, domainfilter.IdentifierTypeHash, tf.IdentifierType)
	assert.Equal(t, []string{"0xabc123"}, tf.Value)
	assert.Equal(t, []domainfilter.TransactionStatus{domainfilter.TransactionStatusFailed}, tf.Statuses)
}

func TestFilterProperties_Map_TransactionFilter_DefaultsToAllStatuses(t *testing.T) {
	fp := &FilterProperties{
		ID:             validFilterID,
		NodeID:         validFilterNodeID,
		Name:           "tx-filter",
		Type:           domainfilter.FilterTypeTransaction.String(),
		IdentifierType: string(domainfilter.IdentifierTypeHash),
		Value:          []string{"0xabc123"},
		// Statuses omitted — Map must default to all statuses
	}

	result, err := fp.Map()

	require.NoError(t, err)
	tf, ok := result.(*domainfilter.TransactionFilter)
	require.True(t, ok)
	assert.Equal(t, domainfilter.AllTransactionStatuses(), tf.Statuses)
}

func TestFilterProperties_Map_EventFilter_AnchorStrategy_Success(t *testing.T) {
	contractAddr := "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	accountAddr := "SomeAccount"
	fp := &FilterProperties{
		ID:              validFilterID,
		NodeID:          validFilterNodeID,
		Name:            "event-filter",
		Type:            domainfilter.FilterTypeEvent.String(),
		Scope:           string(domainfilter.ScopeContract),
		ContractAddress: &contractAddr,
		Specification: FilterSpecificationProperties{
			Strategy:       solana.StrategyAnchor.String(),
			Source:         string(solana.AnchorSourceEmitLog),
			AccountAddress: &accountAddr,
			Signature:      "SomeEventName()",
		},
	}

	result, err := fp.Map()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, domainfilter.FilterTypeEvent, result.Type())
	assert.Equal(t, domainfilter.Name("event-filter"), result.Name())
	ef, ok := result.(*domainfilter.EventFilter)
	require.True(t, ok)
	assert.Equal(t, domainfilter.ScopeContract, ef.Scope)
	assert.Equal(t, solana.StrategyAnchor, ef.Specification.Strategy())
	assert.Equal(t, event.AllContractEventStatuses(), ef.Statuses)
}

func TestFilterProperties_Map_EventFilter_AnchorStrategy_SignatureAndStatuses(t *testing.T) {
	contractAddr := "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	fp := &FilterProperties{
		ID:              validFilterID,
		NodeID:          validFilterNodeID,
		Name:            "event-filter",
		Type:            domainfilter.FilterTypeEvent.String(),
		Scope:           string(domainfilter.ScopeContract),
		ContractAddress: &contractAddr,
		Statuses:        []string{event.ContractEventStatusConfirmed.String(), event.ContractEventStatusFinalized.String()},
		Specification: FilterSpecificationProperties{
			Strategy:       solana.StrategyAnchor.String(),
			Source:         string(solana.AnchorSourceEmitLog),
			AccountAddress: ptr("SomeAccount"),
			Signature:      "Transfer(uint16,string,bytes)",
		},
	}

	result, err := fp.Map()

	require.NoError(t, err)
	ef, ok := result.(*domainfilter.EventFilter)
	require.True(t, ok)
	assert.Equal(t, []event.ContractEventStatus{event.ContractEventStatusConfirmed, event.ContractEventStatusFinalized}, ef.Statuses)
	spec, ok := ef.Specification.(*solana.AnchorSpecification)
	require.True(t, ok)
	assert.Equal(t, "Transfer", spec.EventName())
	require.Len(t, spec.Parameters, 3)
	uintDef, ok := spec.Parameters[0].(solana.UintParameterDefinition)
	require.True(t, ok)
	assert.Equal(t, 16, uintDef.BitSize)
	assert.IsType(t, solana.StringParameterDefinition{}, spec.Parameters[1])
	assert.IsType(t, solana.BytesParameterDefinition{}, spec.Parameters[2])
}

func TestFilterProperties_Map_EventFilter_AnchorStrategy_InvalidSignature(t *testing.T) {
	fp := &FilterProperties{
		ID:     validFilterID,
		NodeID: validFilterNodeID,
		Name:   "event-filter",
		Type:   domainfilter.FilterTypeEvent.String(),
		Scope:  string(domainfilter.ScopeGlobal),
		Specification: FilterSpecificationProperties{
			Strategy:       solana.StrategyAnchor.String(),
			Source:         string(solana.AnchorSourceEmitLog),
			AccountAddress: ptr("SomeAccount"),
			Signature:      "Transfer", // missing "(...)"
		},
	}

	_, err := fp.Map()

	assert.ErrorContains(t, err, "invalid signature")
}

func TestFilterProperties_Map_EventFilter_AnchorStrategy_CompoundParameterTypeAccepted(t *testing.T) {
	fp := &FilterProperties{
		ID:     validFilterID,
		NodeID: validFilterNodeID,
		Name:   "event-filter",
		Type:   domainfilter.FilterTypeEvent.String(),
		Scope:  string(domainfilter.ScopeGlobal),
		Specification: FilterSpecificationProperties{
			Strategy:       solana.StrategyAnchor.String(),
			Source:         string(solana.AnchorSourceEmitLog),
			AccountAddress: ptr("SomeAccount"),
			Signature:      "Transfer(uint16[])",
		},
	}

	result, err := fp.Map()

	require.NoError(t, err)
	ef, ok := result.(*domainfilter.EventFilter)
	require.True(t, ok)
	spec, ok := ef.Specification.(*solana.AnchorSpecification)
	require.True(t, ok)
	require.Len(t, spec.Parameters, 1)
	assert.IsType(t, solana.ArrayParameterDefinition{}, spec.Parameters[0])
}

func TestFilterProperties_Map_EventFilter_SplNativeStrategy_Success(t *testing.T) {
	mintAddress := "SomeMint"
	fp := &FilterProperties{
		ID:     validFilterID,
		NodeID: validFilterNodeID,
		Name:   "event-filter",
		Type:   domainfilter.FilterTypeEvent.String(),
		Scope:  string(domainfilter.ScopeGlobal),
		Specification: FilterSpecificationProperties{
			Strategy:    solana.StrategySplNative.String(),
			Instruction: solana.SplInstructionTransfer,
			MintAddress: &mintAddress,
		},
	}

	result, err := fp.Map()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, domainfilter.FilterTypeEvent, result.Type())
	ef, ok := result.(*domainfilter.EventFilter)
	require.True(t, ok)
	assert.Equal(t, domainfilter.ScopeGlobal, ef.Scope)
	assert.Equal(t, solana.StrategySplNative, ef.Specification.Strategy())
}

func TestFilterProperties_Map_EventFilter_AnchorStrategy_InvalidSource(t *testing.T) {
	// NewAnchorSpecification rejects an unknown source; the error is propagated
	// through NewEventFilter (nil spec → validation failure).
	fp := &FilterProperties{
		ID:     validFilterID,
		NodeID: validFilterNodeID,
		Name:   "event-filter",
		Type:   domainfilter.FilterTypeEvent.String(),
		Scope:  string(domainfilter.ScopeGlobal),
		Specification: FilterSpecificationProperties{
			Strategy:       solana.StrategyAnchor.String(),
			Source:         "INVALID_SOURCE",
			AccountAddress: ptr("SomeAccount"),
			Signature:      "SomeEventName()",
		},
	}

	_, err := fp.Map()

	assert.Error(t, err)
}

func TestFilterProperties_Map_EventFilter_SplNativeStrategy_UnknownInstruction(t *testing.T) {
	// NewSplNativeSpecification rejects an unknown instruction; the error is
	// propagated through NewEventFilter (nil spec → validation failure).
	mintAddress := "SomeMint"
	fp := &FilterProperties{
		ID:     validFilterID,
		NodeID: validFilterNodeID,
		Name:   "event-filter",
		Type:   domainfilter.FilterTypeEvent.String(),
		Scope:  string(domainfilter.ScopeGlobal),
		Specification: FilterSpecificationProperties{
			Strategy:    solana.StrategySplNative.String(),
			Instruction: "unknownInstruction",
			MintAddress: &mintAddress,
		},
	}

	_, err := fp.Map()

	assert.Error(t, err)
}
