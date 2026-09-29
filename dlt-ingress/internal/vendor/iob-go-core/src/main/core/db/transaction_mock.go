package db

import (
	"github.com/stretchr/testify/mock"
)

type TransactionMock struct {
	mock.Mock
}

func (m *TransactionMock) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *TransactionMock) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

var _ Transaction = (*TransactionMock)(nil)
