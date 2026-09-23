package transanctionsender

import "context"

type Port interface {
	SendTransaction(ctx context.Context, request SendTransactionRequest) (*SendTransactionResponse, error)
}
type SendTransactionRequest struct {
	SignedTransaction string
	Dlt               string
	NetworkId         string
}
type SendTransactionResponse struct {
	TxId string
}
