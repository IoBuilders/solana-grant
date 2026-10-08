package retrysvmtransaction

import "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"

type Request struct {
	// Signature of the transaction to retry.
	// It must reference a transaction already stored in the system.
	TxHash string `json:"txHash" binding:"required" swaggertype:"string" example:"5s9f1F8y8n2q1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6"`

	// Compute unit limit to use for the retried transaction.
	// If omitted, the compute unit limit from the original transaction is reused.
	// Applied as given: it is not validated against the network ceiling, so a value above what
	// the target network accepts is rejected by the relay when the transaction is sent.
	CuLimit *amount.Amount `json:"cuLimit" example:"200000"`

	// Compute unit price (priority fee, in micro-lamports per compute unit) to use for the
	// retried transaction. If omitted, the value from the original transaction is reused.
	CuPrice *amount.Amount `json:"cuPrice" example:"5000"`

	// Whether to apply the configured retry fee multiplier to the original compute unit price.
	// Ignored when cuPrice is provided. Default: false.
	UseMultiplier bool `json:"useMultiplier" example:"false" default:"false"`
} // @name RetrySvmTransactionRequest
