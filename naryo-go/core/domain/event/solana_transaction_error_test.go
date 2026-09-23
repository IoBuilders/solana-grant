//go:build test

package event

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustSolanaTxWithErr(t *testing.T, instructions []InstructionData, logs []string, errPayload *string) SolanaTransactionEvent {
	t.Helper()
	tx, err := NewSolanaTransactionEvent(uuid.New(), "sig-1", 1, instructions, nil, logs, errPayload)
	require.NoError(t, err)
	return tx
}

func strPtr(s string) *string { return &s }

func TestSolanaTransactionEvent_DecodedError_NoError(t *testing.T) {
	tx := mustSolanaTxWithErr(t, nil, nil, nil)

	reason, ok := tx.DecodedError(nil)

	assert.False(t, ok)
	assert.Empty(t, reason)
}

func TestSolanaTransactionEvent_DecodedError_BareTopLevelVariant(t *testing.T) {
	tx := mustSolanaTxWithErr(t, nil, nil, strPtr(`"AccountInUse"`))

	reason, ok := tx.DecodedError(nil)

	assert.True(t, ok)
	assert.Equal(t, "AccountInUse", reason)
}

func TestSolanaTransactionEvent_DecodedError_InstructionErrorBareVariant(t *testing.T) {
	tx := mustSolanaTxWithErr(t, nil, nil, strPtr(`{"InstructionError":[1,"InsufficientFunds"]}`))

	reason, ok := tx.DecodedError(nil)

	assert.True(t, ok)
	assert.Equal(t, "InsufficientFunds", reason)
}

func TestSolanaTransactionEvent_DecodedError_CustomResolvedFromAnchorLog(t *testing.T) {
	logs := []string{
		"Program log: AnchorError occurred. Error Code: NotSettlementOperator. Error Number: 6002. Error Message: Only the settlement operator can call this instruction.",
	}
	tx := mustSolanaTxWithErr(t, nil, logs, strPtr(`{"InstructionError":[2,{"Custom":6002}]}`))

	reason, ok := tx.DecodedError(nil)

	assert.True(t, ok)
	assert.Equal(t, "NotSettlementOperator: Only the settlement operator can call this instruction", reason)
}

func TestSolanaTransactionEvent_DecodedError_CustomResolvedFromRegistry_WhenLogMissing(t *testing.T) {
	instructions := []InstructionData{{ProgramID: "program-1"}}
	tx := mustSolanaTxWithErr(t, instructions, nil, strPtr(`{"InstructionError":[0,{"Custom":6010}]}`))
	registry := ProgramErrorRegistry{"program-1": {6010: "InsufficientBalance"}}

	reason, ok := tx.DecodedError(registry)

	assert.True(t, ok)
	assert.Equal(t, "InsufficientBalance", reason)
}

func TestSolanaTransactionEvent_DecodedError_CustomWithNoLogAndNoRegistryEntry(t *testing.T) {
	instructions := []InstructionData{{ProgramID: "program-1"}}
	tx := mustSolanaTxWithErr(t, instructions, nil, strPtr(`{"InstructionError":[0,{"Custom":6010}]}`))

	reason, ok := tx.DecodedError(nil)

	assert.True(t, ok)
	assert.Equal(t, "custom program error: 6010", reason)
}

func TestSolanaTransactionEvent_DecodedError_CustomWithOutOfRangeInstructionIndex(t *testing.T) {
	tx := mustSolanaTxWithErr(t, nil, nil, strPtr(`{"InstructionError":[5,{"Custom":6010}]}`))
	registry := ProgramErrorRegistry{"program-1": {6010: "InsufficientBalance"}}

	reason, ok := tx.DecodedError(registry)

	assert.True(t, ok)
	assert.Equal(t, "custom program error: 6010", reason)
}

func TestSolanaTransactionEvent_DecodedError_MalformedErrJSON_SurfacesRawValue(t *testing.T) {
	tx := mustSolanaTxWithErr(t, nil, nil, strPtr(`not-json`))

	reason, ok := tx.DecodedError(nil)

	assert.True(t, ok)
	assert.Equal(t, "not-json", reason)
}

func TestSolanaTransactionEvent_DecodedError_UnrecognizedInstructionErrorShape_SurfacesRawValue(t *testing.T) {
	raw := `{"InstructionError":[0,{"Something":1}]}`
	tx := mustSolanaTxWithErr(t, nil, nil, strPtr(raw))

	reason, ok := tx.DecodedError(nil)

	assert.True(t, ok)
	assert.Equal(t, raw, reason)
}
