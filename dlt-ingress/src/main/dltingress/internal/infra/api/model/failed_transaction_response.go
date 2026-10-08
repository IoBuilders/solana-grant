package model

import (
	"time"
)

// FailedTransactionResponse represents a transaction that has failed in the DLT.
type FailedTransactionResponse struct {
	TxId         string    `json:"txId" example:"'0xb5c8bd9430b6cc87a0e2fe110ece6bf527fa4f170a4bc8cd032f768fc5219838''"`
	NetworkId    string    `json:"networkId" example:"default"`
	CreatedAt    time.Time `json:"createdAt" format:"date-time" example:"2026-01-01T10:00:00Z"`
	ErrorDetails string    `json:"ErrorDetails,omitempty"`
} // @name FailedTransactionResponse
