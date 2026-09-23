package filter

// TransactionStatus is the on-chain status a TransactionFilter matches
// against. Failed transactions are the primary use case.
type TransactionStatus string

const (
	TransactionStatusFailed      TransactionStatus = "FAILED"
	TransactionStatusConfirmed   TransactionStatus = "CONFIRMED"
	TransactionStatusUnconfirmed TransactionStatus = "UNCONFIRMED"
)

func (s TransactionStatus) IsValid() bool {
	switch s {
	case TransactionStatusFailed, TransactionStatusConfirmed, TransactionStatusUnconfirmed:
		return true
	}
	return false
}

func (s TransactionStatus) String() string {
	return string(s)
}

func AllTransactionStatuses() []TransactionStatus {
	return []TransactionStatus{
		TransactionStatusFailed,
		TransactionStatusConfirmed,
		TransactionStatusUnconfirmed,
	}
}
