//go:build test

package contracttransactionbuilder

import (
	"dlt-ingress/src/main/dltingress/port/contracttransactionbuilder"
	"encoding/base64"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

const builderIdlJSON = `{
  "address": "11111111111111111111111111111112",
  "metadata": { "name": "sample", "version": "0.1.0", "spec": "0.1.0" },
  "instructions": [
    {
      "name": "transfer",
      "discriminator": [1, 2, 3, 4, 5, 6, 7, 8],
      "accounts": [
        { "name": "user",          "writable": true, "signer": true },
        { "name": "recipient",     "writable": true },
        { "name": "systemProgram", "address": "11111111111111111111111111111111" }
      ],
      "args": [
        { "name": "amount", "type": "u64" },
        { "name": "memo",   "type": "string" }
      ]
    }
  ]
}`

const builderBlockhash = "11111111111111111111111111111111"

func newBuilder(t *testing.T) *contracttransactionbuilder.SvmIdlTransactionBuilder {
	t.Helper()
	b, err := contracttransactionbuilder.NewSvmIdlTransactionBuilder([]byte(builderIdlJSON))
	require.NoError(t, err)
	return b
}

func TestDltIngressSVMBuilder_New_InvalidIdlAddress(t *testing.T) {
	badIdl := `{"address":"not-base58!!!","metadata":{"name":"x","version":"0","spec":"0"},"instructions":[]}`
	_, err := contracttransactionbuilder.NewSvmIdlTransactionBuilder([]byte(badIdl))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid IDL address")
}

func TestDltIngressSVMBuilder_New_ParseError(t *testing.T) {
	_, err := contracttransactionbuilder.NewSvmIdlTransactionBuilder([]byte("not json"))
	require.Error(t, err)
}

func TestDltIngressSVMBuilder_Build_PopulatesResponse(t *testing.T) {
	b := newBuilder(t)
	sender := solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	recipient := solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")

	req := contracttransactionbuilder.NewSVMBuildTransactionRequest(
		sender.String(),
		"11111111111111111111111111111112",
		"transfer",
		map[string]any{
			"recipient": recipient.String(),
			"amount":    uint64(42),
			"memo":      "hi",
		},
		builderBlockhash,
		nil,
		nil,
	)

	resp, err := b.BuildTransaction(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.SVMTransactionResponse)
	assert.Nil(t, resp.EVMTransactionResponse)

	svm := resp.SVMTransactionResponse
	assert.Equal(t, sender.String(), svm.FeePayer)
	assert.Equal(t, builderBlockhash, svm.RecentBlockhash)
	assert.Equal(t, uint8(1), svm.Header.NumRequiredSignatures)

	// sender, recipient, programID (target program), systemProgram → 4 keys.
	assert.Len(t, svm.AccountKeys, 4)
	assert.Equal(t, sender.String(), svm.AccountKeys[0]) // signer always first

	require.Len(t, svm.Instructions, 1)
	ix := svm.Instructions[0]
	// 8-byte discriminator + 8-byte u64 LE (42) + 4-byte u32 length (2) + 2 chars = 22 bytes.
	dataBytes, err := base64.StdEncoding.DecodeString(ix.Data)
	require.NoError(t, err)
	require.Len(t, dataBytes, 22)
	assert.Equal(t, []byte{1, 2, 3, 4, 5, 6, 7, 8}, dataBytes[:8])
	assert.Equal(t, []byte{42, 0, 0, 0, 0, 0, 0, 0}, dataBytes[8:16])
	assert.Equal(t, []byte{2, 0, 0, 0, 'h', 'i'}, dataBytes[16:])
}

func TestDltIngressSVMBuilder_Build_RoundTrips_ThroughSolanaTransactionFromBase64(t *testing.T) {
	b := newBuilder(t)
	sender := solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	recipient := solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")

	req := contracttransactionbuilder.NewSVMBuildTransactionRequest(
		sender.String(),
		"11111111111111111111111111111112",
		"transfer",
		map[string]any{"recipient": recipient.String(), "amount": uint64(1), "memo": "x"},
		builderBlockhash,
		nil,
		nil,
	)
	resp, err := b.BuildTransaction(req)
	require.NoError(t, err)

	// Decode the serialized transaction back to confirm it's a well-formed Solana transaction.
	tx, err := solana.TransactionFromBase64(resp.SVMTransactionResponse.SerializedTransaction)
	require.NoError(t, err)

	assert.Equal(t, sender, tx.Message.AccountKeys[0])
	assert.Equal(t, uint8(1), tx.Message.Header.NumRequiredSignatures)
	require.Len(t, tx.Message.Instructions, 1)
}

func TestDltIngressSVMBuilder_Build_IgnoresSmartContractId(t *testing.T) {
	// In SVM the program ID is sourced from the IDL, so req.SmartContractId
	// is informational at this layer — it does not influence routing or the
	// built transaction. Same input args, different SmartContractId values:
	// all should produce a valid transaction targeting idl.Address.
	b := newBuilder(t)
	recipient := solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")
	args := map[string]any{"recipient": recipient.String(), "amount": uint64(1), "memo": "x"}

	cases := []struct {
		name            string
		smartContractId string
	}{
		{"empty", ""},
		{"matches idl", "11111111111111111111111111111112"},
		{"different from idl", "Vote111111111111111111111111111111111111111"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := contracttransactionbuilder.NewSVMBuildTransactionRequest(
				"So11111111111111111111111111111111111111112",
				c.smartContractId,
				"transfer",
				args,
				builderBlockhash,
				nil,
				nil,
			)
			resp, err := b.BuildTransaction(req)
			require.NoError(t, err)
			require.NotNil(t, resp.SVMTransactionResponse)
			// The program account key in the compiled message is always
			// idl.Address, regardless of what req.SmartContractId says.
			assert.Contains(t, resp.SVMTransactionResponse.AccountKeys,
				"11111111111111111111111111111112")
		})
	}
}

func TestDltIngressSVMBuilder_Build_MissingSVMRequest(t *testing.T) {
	b := newBuilder(t)
	req := &contracttransactionbuilder.BuildTransactionRequest{
		SenderDltAccountId: "So11111111111111111111111111111111111111112",
		MethodName:         "transfer",
		MethodArgs:         map[string]any{},
	}
	_, err := b.BuildTransaction(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SVMBuildTransactionRequest is required")
}

func TestDltIngressSVMBuilder_Build_UnknownInstruction(t *testing.T) {
	b := newBuilder(t)
	req := contracttransactionbuilder.NewSVMBuildTransactionRequest(
		"So11111111111111111111111111111111111111112",
		"11111111111111111111111111111112",
		"nope",
		map[string]any{},
		builderBlockhash,
		nil,
		nil,
	)
	_, err := b.BuildTransaction(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `instruction "nope" not found`)
}

func TestDltIngressSVMBuilder_Build_InvalidSender(t *testing.T) {
	b := newBuilder(t)
	req := contracttransactionbuilder.NewSVMBuildTransactionRequest(
		"not-a-pubkey",
		"11111111111111111111111111111112",
		"transfer",
		map[string]any{},
		builderBlockhash,
		nil,
		nil,
	)
	_, err := b.BuildTransaction(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sender")
}

func TestDltIngressSVMBuilder_Build_InvalidBlockhash(t *testing.T) {
	b := newBuilder(t)
	recipient := solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")
	req := contracttransactionbuilder.NewSVMBuildTransactionRequest(
		"So11111111111111111111111111111111111111112",
		"11111111111111111111111111111112",
		"transfer",
		map[string]any{"recipient": recipient.String(), "amount": uint64(1), "memo": "x"},
		"not-a-hash!!!",
		nil,
		nil,
	)
	_, err := b.BuildTransaction(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid recent blockhash")
}

func TestDltIngressSVMBuilder_Build_MissingArg(t *testing.T) {
	b := newBuilder(t)
	recipient := solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")
	req := contracttransactionbuilder.NewSVMBuildTransactionRequest(
		"So11111111111111111111111111111111111111112",
		"11111111111111111111111111111112",
		"transfer",
		map[string]any{"recipient": recipient.String(), "amount": uint64(1)}, // memo missing
		builderBlockhash,
		nil,
		nil,
	)
	_, err := b.BuildTransaction(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `missing arg "memo"`)
}

func TestDltIngressSVMBuilder_Build_PrependsTwoComputeBudgetInstructions_WhenCuLimitAndPriceSet(t *testing.T) {
	b := newBuilder(t)
	sender := solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	recipient := solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")

	cuLimit, _ := amount.NewFromString("200000")
	cuPrice, _ := amount.NewFromString("5000")

	req := contracttransactionbuilder.NewSVMBuildTransactionRequest(
		sender.String(), "11111111111111111111111111111112", "transfer",
		map[string]any{"recipient": recipient.String(), "amount": uint64(1), "memo": "x"},
		builderBlockhash, cuLimit, cuPrice,
	)

	resp, err := b.BuildTransaction(req)
	require.NoError(t, err)

	tx, err := solana.TransactionFromBase64(resp.SVMTransactionResponse.SerializedTransaction)
	require.NoError(t, err)

	// Two compute budget instructions (SetComputeUnitLimit + SetComputeUnitPrice) + 1 program instruction.
	require.Len(t, tx.Message.Instructions, 3)

	computeBudgetProgram := solana.ComputeBudget

	// Instruction 0: SetComputeUnitLimit — 5 bytes total (1 opcode + 4 u32 LE).
	ix0 := tx.Message.Instructions[0]
	assert.Equal(t, computeBudgetProgram, tx.Message.AccountKeys[ix0.ProgramIDIndex])
	assert.Equal(t, computebudget.Instruction_SetComputeUnitLimit, []byte(ix0.Data)[0])

	// Instruction 1: SetComputeUnitPrice — 9 bytes total (1 opcode + 8 u64 LE).
	ix1 := tx.Message.Instructions[1]
	assert.Equal(t, computeBudgetProgram, tx.Message.AccountKeys[ix1.ProgramIDIndex])
	assert.Equal(t, computebudget.Instruction_SetComputeUnitPrice, []byte(ix1.Data)[0])
}

func TestDltIngressSVMBuilder_ImplementsPortInterface(t *testing.T) {
	var _ contracttransactionbuilder.Port = (*contracttransactionbuilder.SvmIdlTransactionBuilder)(nil)
}

func TestDltIngressSVMBuilder_OverrideGasLimit_SetsNewLimit(t *testing.T) {
	b := newBuilder(t)
	sender := solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	recipient := solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")

	req := contracttransactionbuilder.NewSVMBuildTransactionRequest(
		sender.String(), "11111111111111111111111111111112", "transfer",
		map[string]any{"recipient": recipient.String(), "amount": uint64(1), "memo": "x"},
		builderBlockhash, nil, nil,
	)
	original, _ := b.BuildTransaction(req)
	originalCuLimit := req.SVMBuildTransactionRequest.CuLimit

	cuLimit, _ := amount.NewFromString("200000")
	updated, err := b.OverrideGasLimit(*req, *original, cuLimit)

	require.NoError(t, err)
	require.NotNil(t, updated.SVMTransactionResponse)
	assert.Equal(t, cuLimit, updated.SVMTransactionResponse.CuLimit)
	assert.Equal(t, originalCuLimit, req.SVMBuildTransactionRequest.CuLimit, "original request must not be mutated")
}
