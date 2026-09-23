//go:build test

package solana

import (
	"math/big"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

func ptr(s string) *string { return &s }

func TestNewAnchorSpecification(t *testing.T) {
	accountAddress := ptr("testAccountAddress")
	eventName := "testEventName"

	t.Run("Valid", func(t *testing.T) {
		spec, err := NewAnchorSpecification(AnchorSourceEmitLog, accountAddress, eventName, nil)
		assert.NoError(t, err)
		assert.Equal(t, StrategyAnchor, spec.Strategy())
		assert.Equal(t, AnchorSourceEmitLog, spec.Source)
		assert.Equal(t, accountAddress, spec.AccountAddress)
		assert.Equal(t, eventName, spec.EventName())
	})

	t.Run("NilAccountAddress", func(t *testing.T) {
		spec, err := NewAnchorSpecification(AnchorSourceEmitLog, nil, eventName, nil)
		require.NoError(t, err)
		assert.Nil(t, spec.AccountAddress)
	})

	t.Run("InvalidSource", func(t *testing.T) {
		_, err := NewAnchorSpecification(AnchorSource("LOGS"), accountAddress, eventName, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("EmptySource", func(t *testing.T) {
		_, err := NewAnchorSpecification("", accountAddress, eventName, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("BlankAccountAddress", func(t *testing.T) {
		_, err := NewAnchorSpecification(AnchorSourceEmitLog, ptr(""), eventName, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("EmptyEventName", func(t *testing.T) {
		_, err := NewAnchorSpecification(AnchorSourceEmitLog, accountAddress, "", nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("DuplicateParameterPosition", func(t *testing.T) {
		uintDef, err := NewUintParameterDefinition(0, 64)
		require.NoError(t, err)
		stringDef, err := NewStringParameterDefinition(0)
		require.NoError(t, err)
		defs := []ParameterDefinition{uintDef, stringDef}
		_, err = NewAnchorSpecification(AnchorSourceEmitLog, accountAddress, eventName, defs)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}

// stubContractEvent is a hand-written ContractEvent test double.
type stubContractEvent struct {
	eventName  string
	status     event.ContractEventStatus
	parameters []parameter.ContractEventParameter
}

func (stubContractEvent) EventType() event.Type                            { return event.TypeContract }
func (stubContractEvent) NodeID() uuid.UUID                                { return uuid.Nil }
func (e stubContractEvent) Parameters() []parameter.ContractEventParameter { return e.parameters }
func (e stubContractEvent) EventName() string                              { return e.eventName }
func (e stubContractEvent) Status() event.ContractEventStatus              { return e.status }

func TestAnchorSpecification_Matches(t *testing.T) {
	amount, err := parameter.NewSolanaUintParameter(0, big.NewInt(100), 64)
	require.NoError(t, err)
	from, err := parameter.NewSolanaPublicKeyParameter(1, "Account1")
	require.NoError(t, err)

	uintDef, err := NewUintParameterDefinition(0, 64)
	require.NoError(t, err)
	publicKeyDef, err := NewPublicKeyParameterDefinition(1)
	require.NoError(t, err)
	defs := []ParameterDefinition{uintDef, publicKeyDef}
	spec, err := NewAnchorSpecification(AnchorSourceEmitLog, ptr("SomeAccount"), "Transfer", defs)
	require.NoError(t, err)

	t.Run("MatchesByNameAndParameters", func(t *testing.T) {
		e := stubContractEvent{eventName: "Transfer", parameters: []parameter.ContractEventParameter{amount, from}}
		assert.True(t, spec.Matches(e))
	})

	t.Run("NameMismatch", func(t *testing.T) {
		e := stubContractEvent{eventName: "Burn", parameters: []parameter.ContractEventParameter{amount, from}}
		assert.False(t, spec.Matches(e))
	})

	t.Run("ParameterCountMismatch", func(t *testing.T) {
		e := stubContractEvent{eventName: "Transfer", parameters: []parameter.ContractEventParameter{amount}}
		assert.False(t, spec.Matches(e))
	})

	t.Run("ParameterTypeMismatch", func(t *testing.T) {
		mismatched, err := parameter.NewSolanaStringParameter(1, "not-a-key")
		require.NoError(t, err)
		e := stubContractEvent{eventName: "Transfer", parameters: []parameter.ContractEventParameter{amount, mismatched}}
		assert.False(t, spec.Matches(e))
	})

	t.Run("DiscriminatorOnlySpecMatchesNoParameters", func(t *testing.T) {
		discriminatorOnly, err := NewAnchorSpecification(AnchorSourceEmitLog, ptr("SomeAccount"), "Ping", nil)
		require.NoError(t, err)
		e := stubContractEvent{eventName: "Ping", parameters: []parameter.ContractEventParameter{}}
		assert.True(t, discriminatorOnly.Matches(e))
	})
}
