package contractcaller

import (
	"context"
	"math/big"
)

type Port interface {
	Call(ctx context.Context, request CallRequest) (*CallResponse, error)
}

type CallRequest struct {
	To          string
	Data        string
	NetworkId   string
	BlockNumber *big.Int
}

type CallResponse struct {
	Data string
}
