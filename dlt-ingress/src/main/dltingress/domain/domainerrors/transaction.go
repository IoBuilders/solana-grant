package domainerrors

import (
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

const (
	ErrorCodeInvalidTransactionType coreerror.ErrorCode = "INVALID_TRANSACTION_TYPE"
)

func NewInvalidTransactionTypeDomainError(transactionType uint) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeInvalidTransactionType,
		fmt.Sprintf("invalid transaction type: %d (must be 0 or 2)", transactionType),
	)
}
