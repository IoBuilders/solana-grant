package solana

import (
	"strings"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
)

// SPL Token / Token-2022 instruction names. The names match the canonical
// jsonParsed "type" field; each carries the leading discriminator byte of the
// raw instruction data (see knownSplInstructions) so instructions can be
// recognized when decoding binary rather than jsonParsed input. The set covers
// the Token-2022 program (spl-token-2022), a superset of the classic SPL Token
// program naryo ingests events from.
//
// Discriminators are taken from spl_token_2022_interface's TokenInstruction
// unpack and must be revisited when Token-2022 gains new instructions.
const (
	// Account and mint lifecycle.
	SplInstructionInitializeMint                = "initializeMint"
	SplInstructionInitializeMint2               = "initializeMint2"
	SplInstructionInitializeAccount             = "initializeAccount"
	SplInstructionInitializeAccount2            = "initializeAccount2"
	SplInstructionInitializeAccount3            = "initializeAccount3"
	SplInstructionInitializeMultisig            = "initializeMultisig"
	SplInstructionInitializeMultisig2           = "initializeMultisig2"
	SplInstructionInitializeImmutableOwner      = "initializeImmutableOwner"
	SplInstructionInitializeMintCloseAuthority  = "initializeMintCloseAuthority"
	SplInstructionInitializeNonTransferableMint = "initializeNonTransferableMint"
	SplInstructionInitializePermanentDelegate   = "initializePermanentDelegate"
	SplInstructionCreateNativeMint              = "createNativeMint"

	// Balance and authority operations.
	SplInstructionTransfer               = "transfer"
	SplInstructionTransferChecked        = "transferChecked"
	SplInstructionApprove                = "approve"
	SplInstructionApproveChecked         = "approveChecked"
	SplInstructionRevoke                 = "revoke"
	SplInstructionSetAuthority           = "setAuthority"
	SplInstructionMintTo                 = "mintTo"
	SplInstructionMintToChecked          = "mintToChecked"
	SplInstructionBurn                   = "burn"
	SplInstructionBurnChecked            = "burnChecked"
	SplInstructionCloseAccount           = "closeAccount"
	SplInstructionFreezeAccount          = "freezeAccount"
	SplInstructionThawAccount            = "thawAccount"
	SplInstructionSyncNative             = "syncNative"
	SplInstructionReallocate             = "reallocate"
	SplInstructionWithdrawExcessLamports = "withdrawExcessLamports"
	SplInstructionUnwrapLamports         = "unwrapLamports"
	SplInstructionBatch                  = "batch"

	// Read-only views.
	SplInstructionGetAccountDataSize = "getAccountDataSize"
	SplInstructionAmountToUiAmount   = "amountToUiAmount"
	SplInstructionUiAmountToAmount   = "uiAmountToAmount"

	// Token-2022 extensions.
	SplInstructionTransferFeeExtension             = "transferFeeExtension"
	SplInstructionConfidentialTransferExtension    = "confidentialTransferExtension"
	SplInstructionConfidentialTransferFeeExtension = "confidentialTransferFeeExtension"
	SplInstructionConfidentialMintBurnExtension    = "confidentialMintBurnExtension"
	SplInstructionDefaultAccountStateExtension     = "defaultAccountStateExtension"
	SplInstructionMemoTransferExtension            = "memoTransferExtension"
	SplInstructionInterestBearingMintExtension     = "interestBearingMintExtension"
	SplInstructionCpiGuardExtension                = "cpiGuardExtension"
	SplInstructionTransferHookExtension            = "transferHookExtension"
	SplInstructionMetadataPointerExtension         = "metadataPointerExtension"
	SplInstructionGroupPointerExtension            = "groupPointerExtension"
	SplInstructionGroupMemberPointerExtension      = "groupMemberPointerExtension"
	SplInstructionScaledUiAmountExtension          = "scaledUiAmountExtension"
	SplInstructionPausableExtension                = "pausableExtension"
	SplInstructionPermissionedBurnExtension        = "permissionedBurnExtension"
)

// splInstruction records how an instruction is recognized in raw
// (non-jsonParsed) instruction data. Discriminator is the leading byte of the
// instruction data. For Token-2022 extension instructions (Extension == true)
// that byte only identifies the extension family; the concrete sub-instruction
// is carried in the following byte, which this specification does not model.
type splInstruction struct {
	Discriminator byte
	Extension     bool
}

// knownSplInstructions maps each instruction name an SplNativeSpecification is
// allowed to target to its raw-data discriminator. Persisting an unrecognized
// instruction would create a filter that never matches, so construction
// rejects it. Discriminators are taken from spl_token_2022_interface's
// TokenInstruction unpack.
var knownSplInstructions = map[string]splInstruction{
	// Account and mint lifecycle.
	SplInstructionInitializeMint:                {Discriminator: 0},
	SplInstructionInitializeMint2:               {Discriminator: 20},
	SplInstructionInitializeAccount:             {Discriminator: 1},
	SplInstructionInitializeAccount2:            {Discriminator: 16},
	SplInstructionInitializeAccount3:            {Discriminator: 18},
	SplInstructionInitializeMultisig:            {Discriminator: 2},
	SplInstructionInitializeMultisig2:           {Discriminator: 19},
	SplInstructionInitializeImmutableOwner:      {Discriminator: 22},
	SplInstructionInitializeMintCloseAuthority:  {Discriminator: 25},
	SplInstructionInitializeNonTransferableMint: {Discriminator: 32},
	SplInstructionInitializePermanentDelegate:   {Discriminator: 35},
	SplInstructionCreateNativeMint:              {Discriminator: 31},

	// Balance and authority operations.
	SplInstructionTransfer:               {Discriminator: 3},
	SplInstructionTransferChecked:        {Discriminator: 12},
	SplInstructionApprove:                {Discriminator: 4},
	SplInstructionApproveChecked:         {Discriminator: 13},
	SplInstructionRevoke:                 {Discriminator: 5},
	SplInstructionSetAuthority:           {Discriminator: 6},
	SplInstructionMintTo:                 {Discriminator: 7},
	SplInstructionMintToChecked:          {Discriminator: 14},
	SplInstructionBurn:                   {Discriminator: 8},
	SplInstructionBurnChecked:            {Discriminator: 15},
	SplInstructionCloseAccount:           {Discriminator: 9},
	SplInstructionFreezeAccount:          {Discriminator: 10},
	SplInstructionThawAccount:            {Discriminator: 11},
	SplInstructionSyncNative:             {Discriminator: 17},
	SplInstructionReallocate:             {Discriminator: 29},
	SplInstructionWithdrawExcessLamports: {Discriminator: 38},
	SplInstructionUnwrapLamports:         {Discriminator: 45},
	SplInstructionBatch:                  {Discriminator: 255},

	// Read-only views.
	SplInstructionGetAccountDataSize: {Discriminator: 21},
	SplInstructionAmountToUiAmount:   {Discriminator: 23},
	SplInstructionUiAmountToAmount:   {Discriminator: 24},

	// Token-2022 extensions. The discriminator identifies the extension family;
	// the concrete sub-instruction is the following byte of the instruction data.
	SplInstructionTransferFeeExtension:             {Discriminator: 26, Extension: true},
	SplInstructionConfidentialTransferExtension:    {Discriminator: 27, Extension: true},
	SplInstructionConfidentialTransferFeeExtension: {Discriminator: 37, Extension: true},
	SplInstructionConfidentialMintBurnExtension:    {Discriminator: 42, Extension: true},
	SplInstructionDefaultAccountStateExtension:     {Discriminator: 28, Extension: true},
	SplInstructionMemoTransferExtension:            {Discriminator: 30, Extension: true},
	SplInstructionInterestBearingMintExtension:     {Discriminator: 33, Extension: true},
	SplInstructionCpiGuardExtension:                {Discriminator: 34, Extension: true},
	SplInstructionTransferHookExtension:            {Discriminator: 36, Extension: true},
	SplInstructionMetadataPointerExtension:         {Discriminator: 39, Extension: true},
	SplInstructionGroupPointerExtension:            {Discriminator: 40, Extension: true},
	SplInstructionGroupMemberPointerExtension:      {Discriminator: 41, Extension: true},
	SplInstructionScaledUiAmountExtension:          {Discriminator: 43, Extension: true},
	SplInstructionPausableExtension:                {Discriminator: 44, Extension: true},
	SplInstructionPermissionedBurnExtension:        {Discriminator: 46, Extension: true},
}

func IsKnownSplInstruction(name string) bool {
	_, ok := knownSplInstructions[name]
	return ok
}

func SplInstructionDiscriminator(name string) (discriminator byte, extension bool, ok bool) {
	instr, ok := knownSplInstructions[name]
	return instr.Discriminator, instr.Extension, ok
}

// SplNativeSpecification targets a single SPL instruction, optionally
// scoped to one mint. MintAddress is nil when the filter should match the
// instruction regardless of mint — the only option for instructions whose
// raw data carries no mint account at all (e.g. the unchecked Transfer;
// see SplNativeDecoder), and a deliberate choice for any other instruction.
type SplNativeSpecification struct {
	Instruction string
	MintAddress *string
}

func NewSplNativeSpecification(instruction string, mintAddress *string) (SplNativeSpecification, error) {
	if !IsKnownSplInstruction(instruction) {
		return SplNativeSpecification{}, domainerrors.NewInvalidFieldError(
			"Instruction", "SplNativeSpecification", "unknown SPL instruction "+instruction,
		)
	}
	if mintAddress != nil && strings.TrimSpace(*mintAddress) == "" {
		return SplNativeSpecification{}, domainerrors.NewEmptyFieldError("MintAddress", "SplNativeSpecification")
	}
	return SplNativeSpecification{Instruction: instruction, MintAddress: mintAddress}, nil
}

func (s SplNativeSpecification) Strategy() filter.Strategy {
	return StrategySplNative
}

// EventName returns the SPL instruction name: SPL native instructions have
// no separate on-chain event name, so the instruction name itself identifies
// what was decoded.
func (s SplNativeSpecification) EventName() string {
	return s.Instruction
}

// Matches always reports false: SPL native instructions are fixed,
// well-known calls (see knownSplInstructions), not named/parameterized
// emitted events, so they are never resolved through EventFilter-based
// routing — that only applies to Anchor-strategy specifications.
func (s SplNativeSpecification) Matches(event.ContractEvent) bool {
	return false
}

var _ filter.Specification = SplNativeSpecification{}
