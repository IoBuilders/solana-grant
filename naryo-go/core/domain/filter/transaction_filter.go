package filter

import (
	"slices"
	"strings"

	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// TransactionFilter matches transactions rather than events: it selects a
// transaction by IdentifierType/Value and narrows to a set of statuses.
// Unlike an EventFilter it carries no Specification or
// SyncState, which is why the two kinds are modeled separately.
type TransactionFilter struct {
	filter
	IdentifierType IdentifierType
	Value          []string
	Statuses       []TransactionStatus
}

func NewTransactionFilter(
	id uuid.UUID,
	name Name,
	nodeID uuid.UUID,
	identifierType IdentifierType,
	value []string,
	statuses []TransactionStatus,
) (*TransactionFilter, error) {
	if len(statuses) == 0 {
		statuses = AllTransactionStatuses()
	}
	normalizedStatuses := make([]TransactionStatus, len(statuses))
	copy(normalizedStatuses, statuses)

	normalizedValue := make([]string, len(value))
	copy(normalizedValue, value)

	f := &TransactionFilter{
		filter:         filter{id: id, name: name, nodeID: nodeID, filterType: FilterTypeTransaction},
		IdentifierType: identifierType,
		Value:          normalizedValue,
		Statuses:       normalizedStatuses,
	}
	if err := f.Validate(); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *TransactionFilter) Validate() error {
	if err := f.filter.validate(); err != nil {
		return err
	}
	if !f.IdentifierType.IsValid() {
		return domainerrors.NewInvalidFieldError(
			"IdentifierType", "TransactionFilter", "unsupported identifier type "+f.IdentifierType.String(),
		)
	}
	if len(f.Value) == 0 {
		return domainerrors.NewEmptyFieldError("Value", "TransactionFilter")
	}
	for _, v := range f.Value {
		if strings.TrimSpace(v) == "" {
			return domainerrors.NewInvalidFieldError("Value", "TransactionFilter", "blank entry")
		}
	}
	if len(f.Statuses) == 0 {
		return domainerrors.NewEmptyFieldError("Statuses", "TransactionFilter")
	}
	for _, status := range f.Statuses {
		if !status.IsValid() {
			return domainerrors.NewInvalidFieldError(
				"Statuses", "TransactionFilter", "unsupported status "+status.String(),
			)
		}
	}
	return nil
}

// Matches reports whether tx satisfies this filter's Statuses and
// IdentifierType/Value.
//   - Statuses: tx must map (via transactionStatusOf) to one of Statuses.
//   - IdentifierTypeHash matches if tx.Id() equals any one of Value's
//     entries.
//   - IdentifierTypeAddresses matches if every one of Value's entries is
//     present in tx's account list (Solana-only: tx must be a
//     SolanaTransaction).
//   - Every other IdentifierType has no Solana mapping yet — Solana has no
//     single from/to address or Hedera-style entity/token concept the way
//     Ethereum and Hedera do — and never matches.
func (f *TransactionFilter) Matches(tx event.BlockTransaction) bool {
	if !slices.Contains(f.Statuses, transactionStatusOf(tx)) {
		return false
	}
	switch f.IdentifierType {
	case IdentifierTypeHash:
		return containsString(f.Value, tx.Id())
	case IdentifierTypeAddresses:
		solanaTx, ok := tx.(event.SolanaTransaction)
		if !ok {
			return false
		}
		return allContained(f.Value, solanaTx.Accounts)
	case IdentifierTypeToAddress, IdentifierTypeFromAddress, IdentifierTypeIdentityID:
		return false
	default:
		return false
	}
}

// transactionStatusOf maps tx's binary HasError() to a TransactionStatus.
// TransactionStatusUnconfirmed is never produced: TransactionEvent has no
// confirmation-finality concept yet.
func transactionStatusOf(tx event.BlockTransaction) TransactionStatus {
	if tx.HasError() {
		return TransactionStatusFailed
	}
	return TransactionStatusConfirmed
}

// containsString reports whether target is present in values.
func containsString(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

// allContained reports whether every entry in values is present in accounts.
func allContained(values, accounts []string) bool {
	for _, v := range values {
		if !containsString(accounts, v) {
			return false
		}
	}
	return true
}
