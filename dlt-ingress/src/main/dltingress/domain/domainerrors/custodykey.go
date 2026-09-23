package domainerrors

import (
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

const (
	ErrorCodeInvalidKeyType                   coreerror.ErrorCode = "INVALID_KEY_TYPE"
	ErrorCodeInvalidStatus                    coreerror.ErrorCode = "INVALID_STATUS"
	ErrorCodeInvalidDlt                       coreerror.ErrorCode = "INVALID_DLT"
	ErrorCodeInvalidCustodyProvider           coreerror.ErrorCode = "INVALID_CUSTODY_PROVIDER"
	ErrorCodeCustodyKeyNotFoundByDltAccountId coreerror.ErrorCode = "CUSTODY_KEY_NOT_FOUND_BY_DLT_ACCOUNT_ID"
)

func NewInvalidKeyTypeDomainError(keyType string) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeInvalidKeyType,
		fmt.Sprintf("invalid key type: %s", keyType),
	)
}

func NewInvalidStatusDomainError(status string) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeInvalidStatus,
		fmt.Sprintf("invalid status: %s", status),
	)
}

func NewInvalidDltDomainError(dlt string) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeInvalidDlt,
		fmt.Sprintf("invalid DLT: %s", dlt),
	)
}

func NewInvalidCustodyProviderDomainError(provider string) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeInvalidCustodyProvider,
		fmt.Sprintf("invalid custody provider: %s", provider),
	)
}

func NewCustodyKeyNotFoundByDltAccountIdDomainError(dltAccountId string) error {
	return coreerror.NewNotFoundDomainError(
		ErrorCodeCustodyKeyNotFoundByDltAccountId,
		fmt.Sprintf("custody key not found for DLT account ID: %s", dltAccountId),
	)
}
