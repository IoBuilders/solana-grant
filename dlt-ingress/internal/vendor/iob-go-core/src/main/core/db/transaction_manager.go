package db

import (
	"context"
)

type TransactionManager interface {
	WithTransactionValue(parent context.Context) context.Context
	WithNewTransactionValue(parent context.Context) context.Context
	TransactionValue(ctx context.Context) Transaction
	HasTransactionValue(ctx context.Context) bool
}
