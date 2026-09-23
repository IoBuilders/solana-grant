//go:build test

package solanadecoder

import (
	"encoding/binary"
	"errors"
	"math/big"
	"testing"

	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/decoder"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter/solana"
)

const (
	acc0    = "Source11111111111111111111111111111111111"
	acc1    = "Destination111111111111111111111111111111"
	acc2    = "Authority1111111111111111111111111111111111"
	acc3    = "Fourth111111111111111111111111111111111111"
	pubkey0 = "EmbeddedPubkey0000000000000000000000000000"
)

// pubkeyBytes returns a deterministic 32-byte value for embedding directly
// into instruction data (as opposed to an account, which is passed as an
// opaque string).
func pubkeyBytes(seed byte) []byte {
	b := make([]byte, 32)
	for i := range b {
		b[i] = seed
	}
	return b
}

func le64(v uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	return b
}

func le16(v uint16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, v)
	return b
}

func newSpec(t *testing.T, instruction, mint string) solana.SplNativeSpecification {
	t.Helper()
	spec, err := solana.NewSplNativeSpecification(instruction, &mint)
	require.NoError(t, err)
	return spec
}

func newSpecNoMint(t *testing.T, instruction string) solana.SplNativeSpecification {
	t.Helper()
	spec, err := solana.NewSplNativeSpecification(instruction, nil)
	require.NoError(t, err)
	return spec
}

func discriminatorOf(t *testing.T, instruction string) byte {
	t.Helper()
	d, _, ok := solana.SplInstructionDiscriminator(instruction)
	require.True(t, ok)
	return d
}

func assertPubkeyParam(t *testing.T, p parameter.ContractEventParameter, position int, value string) {
	t.Helper()
	require.Equal(t, parameter.TypePublicKey, p.Type())
	assert.Equal(t, position, p.Position())
	assert.Equal(t, value, p.Value())
}

func assertUintParam(t *testing.T, p parameter.ContractEventParameter, position, bitSize int, value uint64) {
	t.Helper()
	require.Equal(t, parameter.TypeUint, p.Type())
	assert.Equal(t, position, p.Position())
	uintP, ok := p.(parameter.SolanaUintParameter)
	require.True(t, ok)
	assert.Equal(t, bitSize, uintP.BitSize)
	assert.Equal(t, value, uintP.Value().(*big.Int).Uint64())
}

func assertPubkeyArrayParam(t *testing.T, p parameter.ContractEventParameter, position int, values ...string) {
	t.Helper()
	require.Equal(t, parameter.TypeArray, p.Type())
	assert.Equal(t, position, p.Position())
	elements := p.Value().([]parameter.ContractEventParameter)
	require.Len(t, elements, len(values))
	for i, v := range values {
		assertPubkeyParam(t, elements[i], i, v)
	}
}

func assertOptionNoneParam(t *testing.T, p parameter.ContractEventParameter, position int) {
	t.Helper()
	require.Equal(t, parameter.TypeOption, p.Type())
	assert.Equal(t, position, p.Position())
	assert.Nil(t, p.Value())
}

func assertOptionSomePubkeyParam(t *testing.T, p parameter.ContractEventParameter, position int, value string) {
	t.Helper()
	require.Equal(t, parameter.TypeOption, p.Type())
	assert.Equal(t, position, p.Position())
	inner, ok := p.Value().(parameter.ContractEventParameter)
	require.True(t, ok)
	assertPubkeyParam(t, inner, position, value)
}

func assertOptionSomeUintParam(t *testing.T, p parameter.ContractEventParameter, position, bitSize int, value uint64) {
	t.Helper()
	require.Equal(t, parameter.TypeOption, p.Type())
	assert.Equal(t, position, p.Position())
	inner, ok := p.Value().(parameter.ContractEventParameter)
	require.True(t, ok)
	assertUintParam(t, inner, position, bitSize, value)
}

func TestSplNativeDecoder_Decode_DiscriminatorMismatch(t *testing.T) {
	spec := newSpec(t, solana.SplInstructionTransfer, pubkey0)
	wrongDiscriminator := discriminatorOf(t, solana.SplInstructionApprove)
	raw := decoder.RawPayload{Data: append([]byte{wrongDiscriminator}, le64(1)...), Accounts: []string{acc0, acc1, acc2}}

	params, err := SplNativeDecoder{}.Decode(spec, raw)

	require.NoError(t, err)
	assert.Nil(t, params)
}

func TestSplNativeDecoder_Decode_EmptyData(t *testing.T) {
	spec := newSpec(t, solana.SplInstructionTransfer, pubkey0)
	raw := decoder.RawPayload{Data: []byte{}, Accounts: []string{acc0, acc1, acc2}}

	params, err := SplNativeDecoder{}.Decode(spec, raw)

	require.NoError(t, err)
	assert.Nil(t, params)
}

func TestSplNativeDecoder_Decode_UnsupportedExtensionInstruction(t *testing.T) {
	spec := newSpec(t, solana.SplInstructionTransferFeeExtension, pubkey0)
	raw := decoder.RawPayload{
		Data:     []byte{discriminatorOf(t, solana.SplInstructionTransferFeeExtension), 0},
		Accounts: []string{acc0},
	}

	params, err := SplNativeDecoder{}.Decode(spec, raw)

	require.Error(t, err)
	assert.True(t, errors.Is(err, domainerrors.ErrValidation))
	assert.Nil(t, params)
}

func TestSplNativeDecoder_Decode_UnsupportedBatchInstruction(t *testing.T) {
	spec := newSpec(t, solana.SplInstructionBatch, pubkey0)
	raw := decoder.RawPayload{Data: []byte{discriminatorOf(t, solana.SplInstructionBatch)}, Accounts: []string{acc0}}

	params, err := SplNativeDecoder{}.Decode(spec, raw)

	require.Error(t, err)
	assert.Nil(t, params)
}

func TestSplNativeDecoder_Decode_WrongSpecificationType(t *testing.T) {
	accountAddress := "acct"
	anchorSpec, err := solana.NewAnchorSpecification(solana.AnchorSourceInstruction, &accountAddress, "someEvent", nil)
	require.NoError(t, err)
	raw := decoder.RawPayload{Data: []byte{discriminatorOf(t, solana.SplInstructionTransfer)}, Accounts: []string{acc0}}

	params, decodeErr := SplNativeDecoder{}.Decode(*anchorSpec, raw)

	require.Error(t, decodeErr)
	assert.Nil(t, params)
}

func TestSplNativeDecoder_Decode_MintAddressMismatch(t *testing.T) {
	spec := newSpec(t, solana.SplInstructionMintTo, "expected-mint")
	raw := decoder.RawPayload{
		Data:     append([]byte{discriminatorOf(t, solana.SplInstructionMintTo)}, le64(100)...),
		Accounts: []string{"different-mint", acc1, acc2},
	}

	params, err := SplNativeDecoder{}.Decode(spec, raw)

	require.NoError(t, err)
	assert.Nil(t, params)
}

func TestSplNativeDecoder_Decode_MintAddressMatch(t *testing.T) {
	spec := newSpec(t, solana.SplInstructionMintTo, acc0)
	raw := decoder.RawPayload{
		Data:     append([]byte{discriminatorOf(t, solana.SplInstructionMintTo)}, le64(100)...),
		Accounts: []string{acc0, acc1, acc2},
	}

	params, err := SplNativeDecoder{}.Decode(spec, raw)

	require.NoError(t, err)
	require.Len(t, params, 5)
	assertUintParam(t, params[3], 3, 64, 100)
}

func TestSplNativeDecoder_Decode_MintAddressIgnoredForMintlessInstruction(t *testing.T) {
	// Transfer carries no mint account at all, so MintAddress can never be
	// verified for it — decode must succeed on discriminator match alone,
	// regardless of what MintAddress was configured to.
	spec := newSpec(t, solana.SplInstructionTransfer, "some-mint-that-cannot-be-checked")
	raw := decoder.RawPayload{
		Data:     append([]byte{discriminatorOf(t, solana.SplInstructionTransfer)}, le64(42)...),
		Accounts: []string{acc0, acc1, acc2},
	}

	params, err := SplNativeDecoder{}.Decode(spec, raw)

	require.NoError(t, err)
	require.Len(t, params, 5)
	assertUintParam(t, params[3], 3, 64, 42)
}

func TestSplNativeDecoder_Decode_NilMintAddress_MatchesAnyMint(t *testing.T) {
	// A nil MintAddress means "match regardless of mint" — this is the only
	// way to build a spec for a mint-less instruction (see the test above),
	// but it's also valid for a mint-carrying instruction that should match
	// any mint: MintTo here decodes even though its Mint account (acc0)
	// isn't checked against anything.
	spec := newSpecNoMint(t, solana.SplInstructionMintTo)
	raw := decoder.RawPayload{
		Data:     append([]byte{discriminatorOf(t, solana.SplInstructionMintTo)}, le64(100)...),
		Accounts: []string{acc0, acc1, acc2},
	}

	params, err := SplNativeDecoder{}.Decode(spec, raw)

	require.NoError(t, err)
	require.Len(t, params, 5)
	assertUintParam(t, params[3], 3, 64, 100)
}

func TestSplNativeDecoder_Decode_AccountsUnderflow(t *testing.T) {
	spec := newSpec(t, solana.SplInstructionTransfer, pubkey0)
	raw := decoder.RawPayload{
		Data:     append([]byte{discriminatorOf(t, solana.SplInstructionTransfer)}, le64(1)...),
		Accounts: []string{acc0, acc1},
	}

	params, err := SplNativeDecoder{}.Decode(spec, raw)

	require.Error(t, err)
	assert.True(t, errors.Is(err, domainerrors.ErrValidation))
	assert.Nil(t, params)
}

func TestSplNativeDecoder_Decode_BufferUnderflow(t *testing.T) {
	spec := newSpec(t, solana.SplInstructionTransfer, pubkey0)
	raw := decoder.RawPayload{
		Data:     []byte{discriminatorOf(t, solana.SplInstructionTransfer), 1, 2, 3},
		Accounts: []string{acc0, acc1, acc2},
	}

	params, err := SplNativeDecoder{}.Decode(spec, raw)

	require.Error(t, err)
	assert.True(t, errors.Is(err, domainerrors.ErrDecoding))
	assert.Nil(t, params)
}

func TestSplNativeDecoder_Decode_InvalidOptionPubkeyTag(t *testing.T) {
	// SetAuthority's newAuthority is an option<pubkey>: a 1-byte tag (0 or 1)
	// followed by a pubkey when Some. 7 is neither.
	spec := newSpec(t, solana.SplInstructionSetAuthority, pubkey0)
	raw := decoder.RawPayload{
		Data:     []byte{discriminatorOf(t, solana.SplInstructionSetAuthority), 2, 7},
		Accounts: []string{acc0, acc2},
	}

	params, err := SplNativeDecoder{}.Decode(spec, raw)

	require.Error(t, err)
	assert.True(t, errors.Is(err, domainerrors.ErrValidation))
	assert.Nil(t, params)
}

func TestSplNativeDecoder_Decode_InvalidOptionUint64Tag(t *testing.T) {
	// UnwrapLamports' amount is an option<u64>: a 1-byte tag (0 or 1)
	// followed by a u64 when Some. 5 is neither.
	spec := newSpec(t, solana.SplInstructionUnwrapLamports, pubkey0)
	raw := decoder.RawPayload{
		Data:     []byte{discriminatorOf(t, solana.SplInstructionUnwrapLamports), 5},
		Accounts: []string{acc0, acc1, acc2},
	}

	params, err := SplNativeDecoder{}.Decode(spec, raw)

	require.Error(t, err)
	assert.True(t, errors.Is(err, domainerrors.ErrValidation))
	assert.Nil(t, params)
}

func TestSplNativeDecoder_Decode_ExtensionTypesArray_OddLength(t *testing.T) {
	// GetAccountDataSize's extensionTypes is a sequence of 2-byte LE values
	// with no length prefix — an odd number of remaining bytes can never be
	// a whole number of them.
	spec := newSpec(t, solana.SplInstructionGetAccountDataSize, acc0)
	raw := decoder.RawPayload{
		Data:     []byte{discriminatorOf(t, solana.SplInstructionGetAccountDataSize), 1, 0, 2},
		Accounts: []string{acc0},
	}

	params, err := SplNativeDecoder{}.Decode(spec, raw)

	require.Error(t, err)
	assert.True(t, errors.Is(err, domainerrors.ErrValidation))
	assert.Nil(t, params)
}

func TestSplNativeDecoder_Decode_PerInstruction(t *testing.T) {
	tests := []struct {
		name     string
		instr    string
		mint     string
		accounts []string
		payload  []byte
		verify   func(t *testing.T, params []parameter.ContractEventParameter)
	}{
		{
			name:     "InitializeMint with freeze authority",
			instr:    solana.SplInstructionInitializeMint,
			mint:     acc0,
			accounts: []string{acc0},
			payload:  append(append([]byte{9}, pubkeyBytes(1)...), append([]byte{1}, pubkeyBytes(2)...)...),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 4)
				assertPubkeyParam(t, params[0], 0, acc0)
				assertUintParam(t, params[1], 1, 8, 9)
				assertPubkeyParam(t, params[2], 2, base58.Encode(pubkeyBytes(1)))
				assertOptionSomePubkeyParam(t, params[3], 3, base58.Encode(pubkeyBytes(2)))
			},
		},
		{
			name:     "InitializeMint2 with no freeze authority",
			instr:    solana.SplInstructionInitializeMint2,
			mint:     acc0,
			accounts: []string{acc0},
			payload:  append(append([]byte{6}, pubkeyBytes(3)...), 0),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 4)
				assertUintParam(t, params[1], 1, 8, 6)
				assertPubkeyParam(t, params[2], 2, base58.Encode(pubkeyBytes(3)))
				assertOptionNoneParam(t, params[3], 3)
			},
		},
		{
			name:     "InitializeAccount",
			instr:    solana.SplInstructionInitializeAccount,
			mint:     acc1,
			accounts: []string{acc0, acc1, acc2},
			payload:  nil,
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 3)
				assertPubkeyParam(t, params[0], 0, acc0)
				assertPubkeyParam(t, params[1], 1, acc1)
				assertPubkeyParam(t, params[2], 2, acc2)
			},
		},
		{
			name:     "InitializeAccount2",
			instr:    solana.SplInstructionInitializeAccount2,
			mint:     acc1,
			accounts: []string{acc0, acc1},
			payload:  pubkeyBytes(5),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 3)
				assertPubkeyParam(t, params[2], 2, base58.Encode(pubkeyBytes(5)))
			},
		},
		{
			name:     "InitializeAccount3",
			instr:    solana.SplInstructionInitializeAccount3,
			mint:     acc1,
			accounts: []string{acc0, acc1},
			payload:  pubkeyBytes(6),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 3)
				assertPubkeyParam(t, params[2], 2, base58.Encode(pubkeyBytes(6)))
			},
		},
		{
			name:     "InitializeMultisig with signers",
			instr:    solana.SplInstructionInitializeMultisig,
			mint:     acc0,
			accounts: []string{acc0, "rent", acc1, acc2},
			payload:  []byte{2},
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 3)
				assertUintParam(t, params[1], 1, 8, 2)
				assertPubkeyArrayParam(t, params[2], 2, acc1, acc2)
			},
		},
		{
			name:     "InitializeMultisig2 with signers",
			instr:    solana.SplInstructionInitializeMultisig2,
			mint:     acc0,
			accounts: []string{acc0, acc1, acc2},
			payload:  []byte{2},
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 3)
				assertPubkeyArrayParam(t, params[2], 2, acc1, acc2)
			},
		},
		{
			name:     "InitializeImmutableOwner",
			instr:    solana.SplInstructionInitializeImmutableOwner,
			mint:     acc0,
			accounts: []string{acc0},
			payload:  nil,
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 1)
				assertPubkeyParam(t, params[0], 0, acc0)
			},
		},
		{
			name:     "InitializeMintCloseAuthority Some",
			instr:    solana.SplInstructionInitializeMintCloseAuthority,
			mint:     acc0,
			accounts: []string{acc0},
			payload:  append([]byte{1}, pubkeyBytes(7)...),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 2)
				assertOptionSomePubkeyParam(t, params[1], 1, base58.Encode(pubkeyBytes(7)))
			},
		},
		{
			name:     "InitializeNonTransferableMint",
			instr:    solana.SplInstructionInitializeNonTransferableMint,
			mint:     acc0,
			accounts: []string{acc0},
			payload:  nil,
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 1)
			},
		},
		{
			name:     "InitializePermanentDelegate",
			instr:    solana.SplInstructionInitializePermanentDelegate,
			mint:     acc0,
			accounts: []string{acc0},
			payload:  pubkeyBytes(8),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 2)
				assertPubkeyParam(t, params[1], 1, base58.Encode(pubkeyBytes(8)))
			},
		},
		{
			name:     "CreateNativeMint",
			instr:    solana.SplInstructionCreateNativeMint,
			mint:     pubkey0,
			accounts: []string{acc0, acc1, acc2},
			payload:  nil,
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 3)
			},
		},
		{
			name:     "Transfer with multisig signers",
			instr:    solana.SplInstructionTransfer,
			mint:     pubkey0,
			accounts: []string{acc0, acc1, acc2, acc3},
			payload:  le64(1000),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 5)
				assertUintParam(t, params[3], 3, 64, 1000)
				assertPubkeyArrayParam(t, params[4], 4, acc3)
			},
		},
		{
			name:     "Approve",
			instr:    solana.SplInstructionApprove,
			mint:     pubkey0,
			accounts: []string{acc0, acc1, acc2},
			payload:  le64(50),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 5)
				assertUintParam(t, params[3], 3, 64, 50)
				assertPubkeyArrayParam(t, params[4], 4)
			},
		},
		{
			name:     "Revoke",
			instr:    solana.SplInstructionRevoke,
			mint:     pubkey0,
			accounts: []string{acc0, acc2},
			payload:  nil,
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 3)
				assertPubkeyArrayParam(t, params[2], 2)
			},
		},
		{
			name:     "SetAuthority",
			instr:    solana.SplInstructionSetAuthority,
			mint:     pubkey0,
			accounts: []string{acc0, acc2},
			payload:  append([]byte{2}, 0),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 5)
				assertUintParam(t, params[2], 2, 8, 2)
				assertOptionNoneParam(t, params[3], 3)
			},
		},
		{
			name:     "MintTo",
			instr:    solana.SplInstructionMintTo,
			mint:     acc0,
			accounts: []string{acc0, acc1, acc2},
			payload:  le64(5),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 5)
				assertUintParam(t, params[3], 3, 64, 5)
			},
		},
		{
			name:     "MintToChecked",
			instr:    solana.SplInstructionMintToChecked,
			mint:     acc0,
			accounts: []string{acc0, acc1, acc2},
			payload:  append(le64(5), 9),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 6)
				assertUintParam(t, params[3], 3, 64, 5)
				assertUintParam(t, params[4], 4, 8, 9)
			},
		},
		{
			name:     "Burn",
			instr:    solana.SplInstructionBurn,
			mint:     acc1,
			accounts: []string{acc0, acc1, acc2},
			payload:  le64(7),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 5)
				assertUintParam(t, params[3], 3, 64, 7)
			},
		},
		{
			name:     "BurnChecked",
			instr:    solana.SplInstructionBurnChecked,
			mint:     acc1,
			accounts: []string{acc0, acc1, acc2},
			payload:  append(le64(7), 6),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 6)
				assertUintParam(t, params[3], 3, 64, 7)
				assertUintParam(t, params[4], 4, 8, 6)
			},
		},
		{
			name:     "CloseAccount",
			instr:    solana.SplInstructionCloseAccount,
			mint:     pubkey0,
			accounts: []string{acc0, acc1, acc2},
			payload:  nil,
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 4)
			},
		},
		{
			name:     "FreezeAccount",
			instr:    solana.SplInstructionFreezeAccount,
			mint:     acc1,
			accounts: []string{acc0, acc1, acc2},
			payload:  nil,
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 4)
			},
		},
		{
			name:     "ThawAccount",
			instr:    solana.SplInstructionThawAccount,
			mint:     acc1,
			accounts: []string{acc0, acc1, acc2},
			payload:  nil,
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 4)
			},
		},
		{
			name:     "TransferChecked",
			instr:    solana.SplInstructionTransferChecked,
			mint:     acc1,
			accounts: []string{acc0, acc1, acc2, acc3},
			payload:  append(le64(1234), 4),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 7)
				assertUintParam(t, params[4], 4, 64, 1234)
				assertUintParam(t, params[5], 5, 8, 4)
			},
		},
		{
			name:     "ApproveChecked",
			instr:    solana.SplInstructionApproveChecked,
			mint:     acc1,
			accounts: []string{acc0, acc1, acc2, acc3},
			payload:  append(le64(1234), 4),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 7)
				assertUintParam(t, params[4], 4, 64, 1234)
				assertUintParam(t, params[5], 5, 8, 4)
			},
		},
		{
			name:     "SyncNative",
			instr:    solana.SplInstructionSyncNative,
			mint:     pubkey0,
			accounts: []string{acc0},
			payload:  nil,
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 1)
			},
		},
		{
			name:     "Reallocate",
			instr:    solana.SplInstructionReallocate,
			mint:     pubkey0,
			accounts: []string{acc0, acc1, acc2, acc3},
			payload:  append(le16(1), le16(3)...),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 6)
				arr, ok := params[4].(parameter.SolanaArrayParameter)
				require.True(t, ok)
				elements := arr.Value().([]parameter.ContractEventParameter)
				require.Len(t, elements, 2)
				assertUintParam(t, elements[0], 0, 16, 1)
				assertUintParam(t, elements[1], 1, 16, 3)
			},
		},
		{
			name:     "WithdrawExcessLamports",
			instr:    solana.SplInstructionWithdrawExcessLamports,
			mint:     pubkey0,
			accounts: []string{acc0, acc1, acc2},
			payload:  nil,
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 4)
			},
		},
		{
			name:     "UnwrapLamports with amount",
			instr:    solana.SplInstructionUnwrapLamports,
			mint:     pubkey0,
			accounts: []string{acc0, acc1, acc2},
			payload:  append([]byte{1}, le64(9)...),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 5)
				assertOptionSomeUintParam(t, params[3], 3, 64, 9)
			},
		},
		{
			name:     "UnwrapLamports without amount",
			instr:    solana.SplInstructionUnwrapLamports,
			mint:     pubkey0,
			accounts: []string{acc0, acc1, acc2},
			payload:  []byte{0},
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 5)
				assertOptionNoneParam(t, params[3], 3)
			},
		},
		{
			name:     "GetAccountDataSize",
			instr:    solana.SplInstructionGetAccountDataSize,
			mint:     acc0,
			accounts: []string{acc0},
			payload:  le16(5),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 2)
				arr, ok := params[1].(parameter.SolanaArrayParameter)
				require.True(t, ok)
				elements := arr.Value().([]parameter.ContractEventParameter)
				require.Len(t, elements, 1)
				assertUintParam(t, elements[0], 0, 16, 5)
			},
		},
		{
			name:     "AmountToUiAmount",
			instr:    solana.SplInstructionAmountToUiAmount,
			mint:     acc0,
			accounts: []string{acc0},
			payload:  le64(100),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 2)
				assertUintParam(t, params[1], 1, 64, 100)
			},
		},
		{
			name:     "UiAmountToAmount",
			instr:    solana.SplInstructionUiAmountToAmount,
			mint:     acc0,
			accounts: []string{acc0},
			payload:  []byte("1.5"),
			verify: func(t *testing.T, params []parameter.ContractEventParameter) {
				require.Len(t, params, 2)
				require.Equal(t, parameter.TypeString, params[1].Type())
				assert.Equal(t, "1.5", params[1].Value())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := newSpec(t, tt.instr, tt.mint)
			raw := decoder.RawPayload{
				Data:     append([]byte{discriminatorOf(t, tt.instr)}, tt.payload...),
				Accounts: tt.accounts,
			}

			params, err := SplNativeDecoder{}.Decode(spec, raw)

			require.NoError(t, err)
			tt.verify(t, params)
		})
	}
}
