//go:build test

package contracttransactionbuilder

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"strings"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
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

func newBuilder(t *testing.T) *SvmIdlTransactionBuilder {
	t.Helper()
	b, err := NewSvmIdlTransactionBuilder([]byte(builderIdlJSON))
	require.NoError(t, err)
	return b
}

func TestDltIngressSVMBuilder_New_InvalidIdlAddress(t *testing.T) {
	badIdl := `{"address":"not-base58!!!","metadata":{"name":"x","version":"0","spec":"0"},"instructions":[]}`
	_, err := NewSvmIdlTransactionBuilder([]byte(badIdl))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid IDL address")
}

func TestDltIngressSVMBuilder_New_ParseError(t *testing.T) {
	_, err := NewSvmIdlTransactionBuilder([]byte("not json"))
	require.Error(t, err)
}

func TestDltIngressSVMBuilder_Build_PopulatesResponse(t *testing.T) {
	b := newBuilder(t)
	sender := solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	recipient := solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")

	req := NewSVMBuildTransactionRequest(
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

	req := NewSVMBuildTransactionRequest(
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

func TestDltIngressSVMBuilder_Build_AcceptsEmptyOrMatchingSmartContractId(t *testing.T) {
	// The program is always the one in the IDL, so an empty id means "that one" and the IDL's own id is
	// accepted too.
	b := newBuilder(t)
	recipient := solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")
	args := map[string]any{"recipient": recipient.String(), "amount": uint64(1), "memo": "x"}

	for name, smartContractId := range map[string]string{
		"empty":       "",
		"matches idl": "11111111111111111111111111111112",
	} {
		t.Run(name, func(t *testing.T) {
			req := NewSVMBuildTransactionRequest(
				"So11111111111111111111111111111111111111112",
				smartContractId,
				"transfer",
				args,
				builderBlockhash,
				nil,
				nil,
			)

			resp, err := b.BuildTransaction(req)

			require.NoError(t, err)
			require.NotNil(t, resp.SVMTransactionResponse)
			assert.Contains(t, resp.SVMTransactionResponse.AccountKeys, "11111111111111111111111111111112")
		})
	}
}

func TestDltIngressSVMBuilder_Build_RejectsASmartContractIdThatIsNotTheIdlProgram(t *testing.T) {
	b := newBuilder(t)
	recipient := solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")
	args := map[string]any{"recipient": recipient.String(), "amount": uint64(1), "memo": "x"}

	for name, smartContractId := range map[string]string{
		"another program":  "Vote111111111111111111111111111111111111111",
		"not a public key": "not-a-key",
	} {
		t.Run(name, func(t *testing.T) {
			req := NewSVMBuildTransactionRequest(
				"So11111111111111111111111111111111111111112",
				smartContractId,
				"transfer",
				args,
				builderBlockhash,
				nil,
				nil,
			)

			resp, err := b.BuildTransaction(req)

			assert.Nil(t, resp)
			var domainErr coreerror.DomainError
			require.True(t, errors.As(err, &domainErr), "expected a domain error, got %v", err)
			assert.Equal(t, domainerrors.ErrorCodeSmartContractIdMismatch, domainErr.ErrorCode())
			assert.Contains(t, err.Error(), smartContractId)
			assert.Contains(t, err.Error(), "11111111111111111111111111111112", "the message names the IDL's program id")
		})
	}
}

func TestDltIngressSVMBuilder_OverrideGasLimit_RejectsASmartContractIdThatIsNotTheIdlProgram(t *testing.T) {
	b := newBuilder(t)
	req := NewSVMBuildTransactionRequest(
		"So11111111111111111111111111111111111111112",
		"Vote111111111111111111111111111111111111111",
		"transfer",
		map[string]any{"recipient": "4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C", "amount": uint64(1), "memo": "x"},
		builderBlockhash,
		nil,
		nil,
	)
	newLimit, err := amount.NewFromString("200000")
	require.NoError(t, err)

	_, err = b.OverrideGasLimit(*req, portcommon.TransactionResponse{}, newLimit)

	var domainErr coreerror.DomainError
	require.True(t, errors.As(err, &domainErr), "expected a domain error, got %v", err)
	assert.Equal(t, domainerrors.ErrorCodeSmartContractIdMismatch, domainErr.ErrorCode())
}

func TestDltIngressSVMBuilder_Build_MissingSVMRequest(t *testing.T) {
	b := newBuilder(t)
	req := &BuildTransactionRequest{
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
	req := NewSVMBuildTransactionRequest(
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
	req := NewSVMBuildTransactionRequest(
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
	req := NewSVMBuildTransactionRequest(
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
	req := NewSVMBuildTransactionRequest(
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

	req := NewSVMBuildTransactionRequest(
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

func TestDltIngressSVMBuilder_Build_RejectsComputeUnitLimitOverflow(t *testing.T) {
	b := newBuilder(t)
	sender := solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	recipient := solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")

	// Above math.MaxUint32 (4,294,967,295): casting straight to uint32 would silently wrap
	// instead of failing, so this must be caught before it reaches the cast.
	cuLimit, _ := amount.NewFromString("4294967296")

	req := NewSVMBuildTransactionRequest(
		sender.String(), "11111111111111111111111111111112", "transfer",
		map[string]any{"recipient": recipient.String(), "amount": uint64(1), "memo": "x"},
		builderBlockhash, nil, cuLimit,
	)

	_, err := b.BuildTransaction(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds the maximum representable value")
}

func TestDltIngressSVMBuilder_Build_RejectsOversizedTransaction(t *testing.T) {
	b := newBuilder(t)
	sender := solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	recipient := solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")

	// A memo well beyond Solana's 1,232-byte wire limit pushes the serialized
	// transaction over the limit on its own.
	oversizedMemo := strings.Repeat("a", 1300)

	req := NewSVMBuildTransactionRequest(
		sender.String(),
		"11111111111111111111111111111112",
		"transfer",
		map[string]any{
			"recipient": recipient.String(),
			"amount":    uint64(1),
			"memo":      oversizedMemo,
		},
		builderBlockhash,
		nil,
		nil,
	)

	_, err := b.BuildTransaction(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds the 1232 byte limit")
}

func TestDltIngressSVMBuilder_ImplementsPortInterface(t *testing.T) {
	var _ Port = (*SvmIdlTransactionBuilder)(nil)
}

func TestDltIngressSVMBuilder_OverrideGasLimit_SetsNewLimit(t *testing.T) {
	b := newBuilder(t)
	sender := solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	recipient := solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")

	req := NewSVMBuildTransactionRequest(
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

func TestDltIngressSVMBuilder_OverrideCuPrice_RebuildsWithNewPrice(t *testing.T) {
	b := newBuilder(t)
	sender := solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	recipient := solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")

	cuLimit, _ := amount.NewFromString("200000")
	req := NewSVMBuildTransactionRequest(
		sender.String(), "11111111111111111111111111111112", "transfer",
		map[string]any{"recipient": recipient.String(), "amount": uint64(1), "memo": "x"},
		builderBlockhash, nil, cuLimit,
	)
	original, err := b.BuildTransaction(req)
	require.NoError(t, err)

	cuPrice, _ := amount.NewFromString("5000")
	updated, err := b.OverrideCuPrice(*req, *original, cuPrice)

	require.NoError(t, err)
	require.NotNil(t, updated.SVMTransactionResponse)
	assert.Equal(t, cuPrice, updated.SVMTransactionResponse.CuPrice)
	assert.Equal(t, cuLimit, updated.SVMTransactionResponse.CuLimit, "the compute unit limit must be kept")
	assert.Nil(t, req.SVMBuildTransactionRequest.CuPrice, "original request must not be mutated")

	tx, err := solana.TransactionFromBase64(updated.SVMTransactionResponse.SerializedTransaction)
	require.NoError(t, err)
	// SetComputeUnitLimit + SetComputeUnitPrice + 1 program instruction.
	require.Len(t, tx.Message.Instructions, 3)
	priceIx := tx.Message.Instructions[1]
	assert.Equal(t, solana.ComputeBudget, tx.Message.AccountKeys[priceIx.ProgramIDIndex])
	assert.Equal(t, computebudget.Instruction_SetComputeUnitPrice, []byte(priceIx.Data)[0])
	assert.Equal(t, uint64(5000), binary.LittleEndian.Uint64(priceIx.Data[1:9]))
}
