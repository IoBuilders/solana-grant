package custody

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
)

type Port interface {
	CreateKey(ctx context.Context, request *CreateKeyRequest) (*KeyResponse, error)
	Sign(ctx context.Context, request *SignRequest) (*SignResponse, error)
}

type CreateKeyRequest struct {
	KeyType string
	Dlt     string
}

type KeyResponse struct {
	DltAccountId string
	ExternalId   string
}

type SignRequest struct {
	Dlt         string
	CustodyKeys []*custodykey.CustodyKey
	Transaction *portcommon.TransactionResponse
}

type SignResponse struct {
	SignedTransaction string
	TxId              string
}
