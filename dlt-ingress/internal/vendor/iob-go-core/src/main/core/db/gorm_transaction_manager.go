package db

import (
	"context"

	"gorm.io/gorm"
)

const TxContextKey = "tx"

type GormTransactionManager struct {
	db *gorm.DB
}

func NewGormTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

func (tm *GormTransactionManager) WithTransactionValue(parent context.Context) context.Context {
	if tx := tm.TransactionValue(parent); tx != nil {
		return parent
	}
	return tm.WithNewTransactionValue(parent)
}

func (tm *GormTransactionManager) WithNewTransactionValue(parent context.Context) context.Context {
	return context.WithValue(parent, TxContextKey, NewGormTransaction(tm.db))
}

func (tm *GormTransactionManager) TransactionValue(ctx context.Context) Transaction {
	val, ok := ctx.Value(TxContextKey).(Transaction)
	if ok {
		return val
	}
	return nil
}

func (tm *GormTransactionManager) HasTransactionValue(ctx context.Context) bool {
	return ctx.Value(TxContextKey) != nil
}
