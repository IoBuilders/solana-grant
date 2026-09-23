package domainerrors

import (
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

const (
	ErrorCodeInvalidNonceValue coreerror.ErrorCode = "INVALID_NONCE_VALUE"
)

func NewInvalidNonceValueDomainError(value *amount.Amount) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeInvalidNonceValue,
		fmt.Sprintf("Invalid nonce value (must be positive and without decimals): %s", value),
	)
}
