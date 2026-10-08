package domainerrors

import (
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

const (
	ErrorCodeInvalidTransactionType          coreerror.ErrorCode = "INVALID_TRANSACTION_TYPE"
	ErrorCodeInvalidNativeTransferArgsLength coreerror.ErrorCode = "INVALID_NATIVE_TRANSFER_ARGS_LENGTH"
	ErrorCodeInvalidNativeTransferArgs       coreerror.ErrorCode = "INVALID_NATIVE_TRANSFER_ARGS"
	ErrorCodeInvalidNativeTransferAmount     coreerror.ErrorCode = "INVALID_NATIVE_TRANSFER_AMOUNT"
	ErrorCodeInvalidNativeTransferTo         coreerror.ErrorCode = "INVALID_NATIVE_TRANSFER_TO"
	ErrorCodeGasLimitExceedsMaximum          coreerror.ErrorCode = "GAS_LIMIT_EXCEEDS_MAXIMUM"
	ErrorCodeNestedCallDepth                 coreerror.ErrorCode = "NESTED_CALL_DEPTH"
	ErrorCodeNestedCallsNotSupported         coreerror.ErrorCode = "NESTED_CALLS_NOT_SUPPORTED"
	ErrorCodeNestedCallsRequired             coreerror.ErrorCode = "NESTED_CALLS_REQUIRED"
	ErrorCodeDispatchCallRequired            coreerror.ErrorCode = "DISPATCH_CALL_REQUIRED"
	ErrorCodeCallDataArgNameRequired         coreerror.ErrorCode = "CALL_DATA_ARG_NAME_REQUIRED"
	ErrorCodeCallDataArgNameTaken            coreerror.ErrorCode = "CALL_DATA_ARG_NAME_TAKEN"
	ErrorCodeNestedCallEncoding              coreerror.ErrorCode = "NESTED_CALL_ENCODING"
	ErrorCodeInvalidStoredSvmTransaction     coreerror.ErrorCode = "INVALID_STORED_SVM_TRANSACTION"
	ErrorCodeComputeUnitLimitOverflow        coreerror.ErrorCode = "COMPUTE_UNIT_LIMIT_OVERFLOW"
	ErrorCodeSvmTransactionSizeExceedsLimit  coreerror.ErrorCode = "SVM_TRANSACTION_SIZE_EXCEEDS_LIMIT"
	ErrorCodeSmartContractIdMismatch         coreerror.ErrorCode = "SMART_CONTRACT_ID_MISMATCH"
	ErrorCodeInvalidSmartContractAddress     coreerror.ErrorCode = "INVALID_SMART_CONTRACT_ADDRESS"
)

func NewNestedCallDepthDomainError(argName string, index int, smartContractName, methodName string) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeNestedCallDepth,
		fmt.Sprintf("nested call %d of %q (%s.%s) carries calls of its own: only one level of nesting is supported",
			index, argName, smartContractName, methodName),
	)
}

func NewInvalidTransactionTypeDomainError(transactionType uint) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeInvalidTransactionType,
		fmt.Sprintf("invalid transaction type: %d (must be 0 or 2)", transactionType),
	)
}

func NewInvalidNativeTransferArgsLength(length int) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeInvalidNativeTransferArgsLength,
		fmt.Sprintf("invalid native transfer args length: %d (must be 2, to and amount)", length),
	)
}

func NewInvalidNativeTransferArgs(args map[string]any) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeInvalidNativeTransferArgs,
		fmt.Sprintf("invalid native transfer args: %v (must be to and amount, all non empty of type string)", args),
	)
}

func NewInvalidNativeTransferAmount(amount string) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeInvalidNativeTransferAmount,
		fmt.Sprintf("invalid native transfer amount: %s (must be a valid big integer number string)", amount),
	)
}

func NewInvalidNativeTransferTo(to string) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeInvalidNativeTransferTo,
		fmt.Sprintf("invalid native transfer to: %s (must be a valid dlt account)", to),
	)
}

// NewGasLimitExceedsMaximumDomainError reports a requested limit above the network's ceiling. The
// wording stays DLT-neutral because it covers both EVM's gas limit and SVM's compute unit limit.
func NewGasLimitExceedsMaximumDomainError(requested, maximum amount.Amount) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeGasLimitExceedsMaximum,
		fmt.Sprintf("requested limit %s exceeds the maximum allowed for this network (%s)", requested.String(), maximum.String()),
	)
}

func NewNestedCallsNotSupportedDomainError(argName string) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeNestedCallsNotSupported,
		fmt.Sprintf("argument %q carries nested calls, which only the sign-and-send-batch command resolves", argName),
	)
}

func NewNestedCallsRequiredDomainError() error {
	return coreerror.NewValidationDomainError(
		ErrorCodeNestedCallsRequired,
		"no calls to batch: sending an empty batch would spend a transaction that does nothing, and a single call is what the sign-and-send command is for",
	)
}

func NewNestedCallEncodingDomainError(argName string, index int, smartContractName, methodName string, cause error) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeNestedCallEncoding,
		fmt.Sprintf("error encoding nested call %d of %q (%s.%s): %s", index, argName, smartContractName, methodName, cause),
	)
}

func NewDispatchCallRequiredDomainError() error {
	return coreerror.NewValidationDomainError(
		ErrorCodeDispatchCallRequired,
		"a dispatch call is required: batching without one means emitting an instruction per call, which no builder supports yet",
	)
}

func NewCallDataArgNameRequiredDomainError() error {
	return coreerror.NewValidationDomainError(
		ErrorCodeCallDataArgNameRequired,
		"CallDataArgName is required: without it there is no argument of the dispatch call to put the batched calls in",
	)
}

func NewCallDataArgNameTakenDomainError(argName string) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeCallDataArgNameTaken,
		fmt.Sprintf("the dispatch call already has an argument named %q, which is reserved for the batched calls", argName),
	)
}

func NewInvalidStoredSvmTransactionDomainError(cause error) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeInvalidStoredSvmTransaction,
		fmt.Sprintf("could not decode stored transaction: %s", cause),
	)
}

func NewComputeUnitLimitOverflowDomainError(requested uint64, maximum uint32) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeComputeUnitLimitOverflow,
		fmt.Sprintf("compute unit limit %d exceeds the maximum representable value (%d)", requested, maximum),
	)
}

func NewInvalidSmartContractAddressDomainError(smartContractId string) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeInvalidSmartContractAddress,
		fmt.Sprintf("smart contract id %q is not a valid EVM address", smartContractId),
	)
}

func NewSmartContractIdMismatchDomainError(smartContractId, programId string) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeSmartContractIdMismatch,
		fmt.Sprintf("smart contract id %q is not the program id %s of the contract's IDL", smartContractId, programId),
	)
}

func NewSvmTransactionSizeExceedsLimitDomainError(size, limit int) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeSvmTransactionSizeExceedsLimit,
		fmt.Sprintf("serialized transaction size %d bytes exceeds the %d byte limit", size, limit),
	)
}
