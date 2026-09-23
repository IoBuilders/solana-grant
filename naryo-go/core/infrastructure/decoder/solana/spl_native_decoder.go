package solanadecoder

import (
	"fmt"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/decoder"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter/solana"
)

// SplNativeDecoder decodes SPL Token / Token-2022 native instructions —
// identified by a single leading discriminator byte (see
// solana.SplInstructionDiscriminator), not an 8-byte Anchor discriminator —
// into ContractEventParameter values. Each instruction has a fixed account
// and data layout defined by the SPL Token program itself, so unlike the
// Anchor path this does not go through DecodeBuffer/ParameterDefinition:
// some fields (extension-type arrays, the UiAmountToAmount string) consume
// the rest of the data buffer with no length prefix, which a generic
// Borsh-style decoder cannot express.
//
// Per-instruction decode functions live alongside this file grouped by
// topic: spl_native_mint.go (mint lifecycle), spl_native_account.go (token
// account lifecycle), spl_native_multisig.go, spl_native_transfer.go
// (balance/authority operations), spl_native_lamports.go. Shared building
// blocks (parameter constructors, option/array decoding) live in
// spl_native_params.go.
type SplNativeDecoder struct{}

var _ decoder.Decoder = SplNativeDecoder{}

// splInstructionHandler describes how to recognize and decode one SPL
// instruction.
type splInstructionHandler struct {
	// mintAccountIndex is the index into RawPayload.Accounts holding the
	// mint, or -1 if this instruction's layout carries no mint account at
	// all. A Specification's MintAddress can only ever be verified against
	// instructions with mintAccountIndex >= 0 — see the MintAddress check
	// in Decode below.
	mintAccountIndex int
	decode           func(data []byte, accounts []string) ([]parameter.ContractEventParameter, error)
}

// splInstructionHandlers covers every non-extension, non-Batch SPL
// instruction. Extension-flagged instructions and Batch are TLV/nested
// encoded and are not decoded here.
var splInstructionHandlers = map[string]splInstructionHandler{
	solana.SplInstructionInitializeMint:                {mintAccountIndex: 0, decode: decodeInitializeMint},
	solana.SplInstructionInitializeMint2:               {mintAccountIndex: 0, decode: decodeInitializeMint},
	solana.SplInstructionInitializeAccount:             {mintAccountIndex: 1, decode: decodeInitializeAccount},
	solana.SplInstructionInitializeAccount2:            {mintAccountIndex: 1, decode: decodeInitializeAccount23},
	solana.SplInstructionInitializeAccount3:            {mintAccountIndex: 1, decode: decodeInitializeAccount23},
	solana.SplInstructionInitializeMultisig:            {mintAccountIndex: -1, decode: decodeInitializeMultisig(1)},
	solana.SplInstructionInitializeMultisig2:           {mintAccountIndex: -1, decode: decodeInitializeMultisig(0)},
	solana.SplInstructionInitializeImmutableOwner:      {mintAccountIndex: -1, decode: decodeSingleAccountNoData},
	solana.SplInstructionInitializeMintCloseAuthority:  {mintAccountIndex: 0, decode: decodeInitializeMintCloseAuthority},
	solana.SplInstructionInitializeNonTransferableMint: {mintAccountIndex: 0, decode: decodeSingleAccountNoData},
	solana.SplInstructionInitializePermanentDelegate:   {mintAccountIndex: 0, decode: decodeInitializePermanentDelegate},
	solana.SplInstructionCreateNativeMint:              {mintAccountIndex: -1, decode: decodeCreateNativeMint},

	solana.SplInstructionTransfer:               {mintAccountIndex: -1, decode: decodeAmountWithAuthority(2)},
	solana.SplInstructionApprove:                {mintAccountIndex: -1, decode: decodeAmountWithAuthority(2)},
	solana.SplInstructionRevoke:                 {mintAccountIndex: -1, decode: decodeRevoke},
	solana.SplInstructionSetAuthority:           {mintAccountIndex: -1, decode: decodeSetAuthority},
	solana.SplInstructionMintTo:                 {mintAccountIndex: 0, decode: decodeAmountWithAuthority(2)},
	solana.SplInstructionMintToChecked:          {mintAccountIndex: 0, decode: decodeAmountCheckedWithAuthority(2)},
	solana.SplInstructionBurn:                   {mintAccountIndex: 1, decode: decodeAmountWithAuthority(2)},
	solana.SplInstructionBurnChecked:            {mintAccountIndex: 1, decode: decodeAmountCheckedWithAuthority(2)},
	solana.SplInstructionCloseAccount:           {mintAccountIndex: -1, decode: decodeCloseAccount},
	solana.SplInstructionFreezeAccount:          {mintAccountIndex: 1, decode: decodeFreezeThawAccount},
	solana.SplInstructionThawAccount:            {mintAccountIndex: 1, decode: decodeFreezeThawAccount},
	solana.SplInstructionSyncNative:             {mintAccountIndex: -1, decode: decodeSingleAccountNoData},
	solana.SplInstructionReallocate:             {mintAccountIndex: -1, decode: decodeReallocate},
	solana.SplInstructionWithdrawExcessLamports: {mintAccountIndex: -1, decode: decodeWithdrawExcessLamports},
	solana.SplInstructionUnwrapLamports:         {mintAccountIndex: -1, decode: decodeUnwrapLamports},

	solana.SplInstructionTransferChecked: {mintAccountIndex: 1, decode: decodeCheckedTransferApprove},
	solana.SplInstructionApproveChecked:  {mintAccountIndex: 1, decode: decodeCheckedTransferApprove},

	solana.SplInstructionGetAccountDataSize: {mintAccountIndex: 0, decode: decodeGetAccountDataSize},
	solana.SplInstructionAmountToUiAmount:   {mintAccountIndex: 0, decode: decodeAmountToUiAmount},
	solana.SplInstructionUiAmountToAmount:   {mintAccountIndex: 0, decode: decodeUiAmountToAmount},
}

// Decode implements decoder.Decoder for SPL_NATIVE-strategy specifications.
func (SplNativeDecoder) Decode(spec filter.Specification, raw decoder.RawPayload) ([]parameter.ContractEventParameter, error) {
	splSpec, ok := spec.(solana.SplNativeSpecification)
	if !ok {
		return nil, domainerrors.NewInvalidFieldError(
			"Specification", "SplNativeDecoder", fmt.Sprintf("expected solana.SplNativeSpecification, got %T", spec),
		)
	}

	if len(raw.Data) < 1 {
		return nil, nil
	}

	discriminator, extension, ok := solana.SplInstructionDiscriminator(splSpec.Instruction)
	if !ok || extension {
		return nil, domainerrors.NewInvalidFieldError(
			"Instruction", "SplNativeDecoder", fmt.Sprintf("SPL instruction %q is not supported for native decoding", splSpec.Instruction),
		)
	}
	if raw.Data[0] != discriminator {
		return nil, nil
	}

	handler, ok := splInstructionHandlers[splSpec.Instruction]
	if !ok {
		return nil, domainerrors.NewInvalidFieldError(
			"Instruction", "SplNativeDecoder", fmt.Sprintf("SPL instruction %q is not supported for native decoding", splSpec.Instruction),
		)
	}

	// MintAddress is only ever checked when both the spec asked for it (it's
	// nilable — see solana.SplNativeSpecification) and the instruction's
	// layout actually carries a mint account to check it against. A spec
	// with MintAddress set against a mint-less instruction (e.g. the
	// unchecked Transfer) matches on discriminator alone: the mint genuinely
	// cannot be verified from the instruction's own bytes.
	if handler.mintAccountIndex >= 0 && splSpec.MintAddress != nil {
		if handler.mintAccountIndex >= len(raw.Accounts) {
			return nil, domainerrors.NewInvalidFieldError(
				"Accounts", "SplNativeDecoder",
				fmt.Sprintf("instruction %q requires at least %d accounts, got %d", splSpec.Instruction, handler.mintAccountIndex+1, len(raw.Accounts)),
			)
		}
		if raw.Accounts[handler.mintAccountIndex] != *splSpec.MintAddress {
			return nil, nil
		}
	}

	return handler.decode(raw.Data[1:], raw.Accounts)
}
