package retryevmtransaction

import "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"

type Request struct {
	// Hash of the transaction to retry.
	// It must reference a transaction already stored in the system.
	TxHash string `json:"txHash" binding:"required" swaggertype:"string"  example:"0x8f6c..."`

	// Gas limit to used for the retried transaction.
	// If omitted, the gas limit from the original transaction is reused.
	// Applied as given: it is not validated against the network ceiling, so a value above what
	// the target network accepts is rejected by the relay when the transaction is sent.
	GasLimit *amount.Amount `json:"gasLimit" example:"15000000"`

	// Gas price to use for the retried transaction.
	// Applicable to legacy transactions (type 0).
	// If omitted, the gas price from the original transaction is reused.
	GasPrice *amount.Amount `json:"gasPrice" example:"20000000000"`

	// Max priority fee per gas to use for the retried transaction.
	// Applicable to EIP-1559 dynamic fee transactions (type 2).
	// If omitted, the value from the original transaction is reused.
	MaxPriorityFeePerGas *amount.Amount `json:"maxPriorityFeePerGas" example:"2000000000"`

	// Whether to apply the configured gas multiplier when retrying the transaction.
	// When false, the provided fee values (or the original ones if not provided) are used as-is.
	// Default: false.
	UseMultiplier bool `json:"useMultiplier" example:"false" default:"false"`

	// Nonce to use for the retried transaction.
	// If omitted, the current address nonce is retrieved.
	Nonce *amount.Amount `json:"nonce" example:"15"`
} // @name RetryEvmTransactionRequest
