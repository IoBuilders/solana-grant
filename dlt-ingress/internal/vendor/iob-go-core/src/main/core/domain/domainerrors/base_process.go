package domainerrors

import (
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

const (
	ErrorCodeBaseProcessNotFound         coreerror.ErrorCode = "BASE_PROCESS_NOT_FOUND"
	ErrorCodeBaseProcessNotFoundByTxHash coreerror.ErrorCode = "BASE_PROCESS_NOT_FOUND_BY_TX_HASH"
)

func NewBaseProcessNotFoundDomainError(id uuid.UUID) error {
	return coreerror.NewNotFoundDomainError(ErrorCodeBaseProcessNotFound, fmt.Sprintf("BaseProcess with id %s not found", id))
}

func NewBaseProcessNotFoundByTxHashDomainError(txHash string) error {
	return coreerror.NewNotFoundDomainError(ErrorCodeBaseProcessNotFoundByTxHash, fmt.Sprintf("BaseProcess with transaction hash %s not found", txHash))
}

const ErrorCodeProcessTypeNotRetryable coreerror.ErrorCode = "PROCESS_TYPE_NOT_RETRYABLE"

func NewProcessTypeNotRetryableDomainError(processType string) error {
	return coreerror.NewConflictDomainError(ErrorCodeProcessTypeNotRetryable,
		fmt.Sprintf("Process type %s has no retry entry point registered in this bounded context", processType))
}
