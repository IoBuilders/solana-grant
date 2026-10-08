package model

type RetryTransactionModel struct {
	// Hash of the newly submitted transaction to the DLT.
	// This value changes if any parameter (e.g., gas price, nonce) is modified.
	NewTxHash string `json:"newTxHash" example:"0xa1b2c3d4e5f607..."`
} // @name RetriedTransaction
