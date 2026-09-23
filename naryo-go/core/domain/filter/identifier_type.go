package filter

// IdentifierType selects which field of a transaction a TransactionFilter
// matches Value against.
type IdentifierType string

const (
	IdentifierTypeHash        IdentifierType = "HASH"
	IdentifierTypeToAddress   IdentifierType = "TO_ADDRESS"
	IdentifierTypeFromAddress IdentifierType = "FROM_ADDRESS"
	IdentifierTypeIdentityID  IdentifierType = "IDENTITY_ID"
	// IdentifierTypeAddresses matches every one of Value's addresses against
	// a transaction's account list (see TransactionFilter.Matches).
	IdentifierTypeAddresses IdentifierType = "ADDRESSES"
)

func (t IdentifierType) IsValid() bool {
	switch t {
	case IdentifierTypeHash, IdentifierTypeToAddress, IdentifierTypeFromAddress, IdentifierTypeIdentityID, IdentifierTypeAddresses:
		return true
	}
	return false
}

func (t IdentifierType) String() string {
	return string(t)
}
