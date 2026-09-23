package svm

import "context"

type BlockhashProvider interface {
	GetRecentBlockhash(ctx context.Context, networkId string) (string, error)
}
