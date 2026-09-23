package nonceprovider

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type Port interface {
	GetNonce(ctx context.Context, request *GetNonceRequest) (*amount.Amount, error)
	SetNonce(ctx context.Context, request *SetNonceRequest) error
}

type GetNonceRequest struct {
	DltAccountId string
	NetworkId    string
}
type SetNonceRequest struct {
	DltAccountId string
	NetworkId    string
	Value        *amount.Amount
}
