package db

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type TransactionManagerMock struct {
	mock.Mock
}

func (m *TransactionManagerMock) WithTransactionValue(parent context.Context) context.Context {
	args := m.Called(parent)
	return args.Get(0).(context.Context)
}

func (m *TransactionManagerMock) WithNewTransactionValue(parent context.Context) context.Context {
	args := m.Called(parent)
	return args.Get(0).(context.Context)
}

func (m *TransactionManagerMock) TransactionValue(ctx context.Context) Transaction {
	args := m.Called(ctx)
	return args.Get(0).(Transaction)
}

func (m *TransactionManagerMock) HasTransactionValue(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Get(0).(bool)
}

var _ TransactionManager = (*TransactionManagerMock)(nil)
