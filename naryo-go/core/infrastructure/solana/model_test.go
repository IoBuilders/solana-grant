//go:build test

package solana

import (
	"encoding/json"
	"testing"

	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testPubkeyA = "4wBqpZM9xaSheZzJSMawUKKwhdpChKbZ5eu5ky4Vigw"
	testPubkeyB = "JEJUoGfGEPTZ1XTwN39dYdFxYxDiDaSKVNy5qYWJmZt3"
	testPubkeyC = "Cs8KY3PiWrCMAytMsBRQo8EdGbticVtdvufLnb2UhXh"
)

func TestAccountsFromParsedInfo(t *testing.T) {
	t.Run("FlatFields", func(t *testing.T) {
		info := json.RawMessage(`{"source": "` + testPubkeyA + `", "destination": "` + testPubkeyB + `", "lamports": 1000}`)
		accounts, err := accountsFromParsedInfo(info)
		require.NoError(t, err)
		assert.Equal(t, []string{testPubkeyA, testPubkeyB}, accounts)
	})

	t.Run("NestedObjectsAndArrays", func(t *testing.T) {
		info := json.RawMessage(`{
			"multisigAuthority": "` + testPubkeyA + `",
			"signers": ["` + testPubkeyB + `", "` + testPubkeyC + `"],
			"tokenAmount": {"amount": "1000000", "decimals": 6, "uiAmountString": "1"}
		}`)
		accounts, err := accountsFromParsedInfo(info)
		require.NoError(t, err)
		assert.Equal(t, []string{testPubkeyA, testPubkeyC, testPubkeyB}, accounts)
	})

	t.Run("DeduplicatesRepeatedPubkeys", func(t *testing.T) {
		info := json.RawMessage(`{"source": "` + testPubkeyA + `", "owner": "` + testPubkeyA + `"}`)
		accounts, err := accountsFromParsedInfo(info)
		require.NoError(t, err)
		assert.Equal(t, []string{testPubkeyA}, accounts)
	})

	t.Run("NoAccounts", func(t *testing.T) {
		info := json.RawMessage(`{"decimals": 6, "mintAuthority": null}`)
		accounts, err := accountsFromParsedInfo(info)
		require.NoError(t, err)
		assert.Empty(t, accounts)
	})

	t.Run("NonPubkeyStringsExcluded", func(t *testing.T) {
		// "authorityType" and a plain decimal amount both look nothing like a
		// pubkey once base58-decoded: a short digit run decodes far shorter
		// than 32 bytes.
		info := json.RawMessage(`{"authorityType": "mintTokens", "amount": "1000000"}`)
		accounts, err := accountsFromParsedInfo(info)
		require.NoError(t, err)
		assert.Empty(t, accounts)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		_, err := accountsFromParsedInfo(json.RawMessage(`not json`))
		assert.Error(t, err)
	})

	t.Run("EmptyInfo", func(t *testing.T) {
		accounts, err := accountsFromParsedInfo(nil)
		require.NoError(t, err)
		assert.Empty(t, accounts)
	})

	t.Run("NullInfo", func(t *testing.T) {
		accounts, err := accountsFromParsedInfo(json.RawMessage(`null`))
		require.NoError(t, err)
		assert.Empty(t, accounts)
	})
}

// A recognized program's "parsed" instruction isn't always an {info: ...}
// object: the SPL Memo program (among others) parses to a bare JSON string
// (the memo text) instead. Real payload captured from a live getBlock call.
const memoInstructionJSON = `{
	"program": "spl-memo",
	"programId": "MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr",
	"parsed": "forge:force-sell:v1:f1688918aea8c86582616ea4c6fc19633640fe65b2014d56512143b4bcfed812000000030000000000000000",
	"stackHeight": 1
}`

func TestRpcInstruction_ParsedIsString_UnmarshalsWithoutError(t *testing.T) {
	var ri rpcInstruction
	err := json.Unmarshal([]byte(memoInstructionJSON), &ri)

	require.NoError(t, err)
	assert.Equal(t, "MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr", ri.ProgramID)
	require.NotNil(t, ri.Parsed)
	assert.Empty(t, ri.Parsed.Info)
}

func TestRpcInstruction_ToDomain_ParsedIsString_ReturnsInstructionWithNoAccounts(t *testing.T) {
	var ri rpcInstruction
	require.NoError(t, json.Unmarshal([]byte(memoInstructionJSON), &ri))

	instr, err := ri.toDomain(false)

	require.NoError(t, err)
	assert.Equal(t, "MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr", instr.ProgramID)
	assert.Empty(t, instr.Accounts)
	assert.Empty(t, instr.Data)
	assert.False(t, instr.Inner)
}

func TestRpcInstruction_ToDomain_Inner_SetsInnerTrue(t *testing.T) {
	var ri rpcInstruction
	require.NoError(t, json.Unmarshal([]byte(memoInstructionJSON), &ri))

	instr, err := ri.toDomain(true)

	require.NoError(t, err)
	assert.True(t, instr.Inner)
}

// Some inner instructions legitimately have no data at all — base58.Decode errors on "" as
// invalid input, so toDomain must not call it for an empty Data field.
func TestRpcInstruction_ToDomain_UnrecognizedProgram_EmptyData_DoesNotError(t *testing.T) {
	ri := rpcInstruction{ProgramID: "prog", Data: "", Accounts: []string{"acc-1"}}

	instr, err := ri.toDomain(true)

	require.NoError(t, err)
	assert.Empty(t, instr.Data)
	assert.Equal(t, []string{"acc-1"}, instr.Accounts)
}

func TestRpcInstruction_ToDomain_UnrecognizedProgram_NonEmptyData_IsDecoded(t *testing.T) {
	ri := rpcInstruction{ProgramID: "prog", Data: "2NEpo7TZRRrLZSi2U", Accounts: []string{"acc-1"}}

	instr, err := ri.toDomain(false)

	require.NoError(t, err)
	decoded, err := base58.Decode("2NEpo7TZRRrLZSi2U")
	require.NoError(t, err)
	assert.Equal(t, decoded, instr.Data)
	assert.NotEmpty(t, instr.Data)
}

// rt.toDomain preallocates capacity from len(Message.Instructions) alone: len(InnerInstructions)
// is the number of inner-instruction *groups*, not the total individual inner instructions each
// group holds, so a transaction with an uneven group size (one group of 1, one of 2, here) is
// exactly what would have exposed a wrong preallocated length as gaps/garbage entries.
func TestRpcTransaction_ToDomain_TopLevelAndInnerInstructions_PreservesAllInOrder(t *testing.T) {
	const raw = `{
		"meta": {
			"err": null,
			"logMessages": [],
			"innerInstructions": [
				{"index": 0, "instructions": [{"programId": "innerProg1", "data": "z", "accounts": []}]},
				{"index": 1, "instructions": [
					{"programId": "innerProg2a", "data": "z", "accounts": []},
					{"programId": "innerProg2b", "data": "z", "accounts": []}
				]}
			]
		},
		"transaction": {
			"signatures": ["sig1"],
			"message": {
				"accountKeys": [],
				"instructions": [
					{"programId": "topProg1", "data": "z", "accounts": []},
					{"programId": "topProg2", "data": "z", "accounts": []}
				]
			}
		}
	}`

	var rt rpcTransaction
	require.NoError(t, json.Unmarshal([]byte(raw), &rt))

	tx, err := rt.toDomain()

	require.NoError(t, err)
	require.Len(t, tx.Instructions, 5)

	var programIDs []string
	var innerFlags []bool
	for _, instr := range tx.Instructions {
		programIDs = append(programIDs, instr.ProgramID)
		innerFlags = append(innerFlags, instr.Inner)
	}
	assert.Equal(t, []string{"topProg1", "topProg2", "innerProg1", "innerProg2a", "innerProg2b"}, programIDs)
	assert.Equal(t, []bool{false, false, true, true, true}, innerFlags)
}
