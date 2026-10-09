package gettransaction

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/evmtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/transaction/evmtransaction/mock"
	"fmt"
	"testing"

	txfactory "dlt-ingress/src/main/dltingress/internal/domain/transaction/evmtransaction"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDltIngressGetTransactionQueryHandler_Execute_Success(t *testing.T) {
	repo := new(mockevmtransactionrepo.Postgres)
	factory := txfactory.NewEvmTransactionTestFactory()
	tx := factory.CreateEntity()

	repo.On("FindByTxId", mock.Anything, tx.TxId).Return(tx, nil)

	handler := NewHandler(repo)
	rawResp, err := handler.Execute(context.Background(), Query{
		TxId: tx.TxId,
		Dlt:  "EVM",
	})

	assert.Nil(t, err)
	resp := rawResp.(Response)
	assert.Equal(t, tx.TxId, resp.TxId)
	assert.Equal(t, tx.NetworkId, resp.NetworkId)
	assert.Equal(t, tx.NetworkUrl, resp.NetworkUrl)
	assert.Equal(t, string(tx.Dlt), resp.Dlt)
	assert.NotNil(t, resp.Evm)
	assert.Equal(t, tx.FromAddress, resp.Evm.FromAddress)
	assert.Equal(t, tx.ToAddress, resp.Evm.ToAddress)
	assert.Equal(t, tx.Nonce, resp.Evm.Nonce)
	assert.Equal(t, tx.TransactionType, resp.Evm.TransactionType)
	assert.Equal(t, tx.Data, resp.Evm.Data)
	repo.AssertExpectations(t)
}

func TestDltIngressGetTransactionQueryHandler_Execute_NotFound(t *testing.T) {
	repo := new(mockevmtransactionrepo.Postgres)
	txId := "0xNotFound"

	repo.On("FindByTxId", mock.Anything, txId).
		Return((*evmtransaction.EvmTransaction)(nil), assert.AnError)

	handler := NewHandler(repo)
	rawResp, err := handler.Execute(context.Background(), Query{
		TxId: txId,
		Dlt:  "EVM",
	})

	assert.Nil(t, rawResp)
	assert.NotNil(t, err)
	repo.AssertExpectations(t)
}

func TestDltIngressGetTransactionQueryHandler_Execute_RepositoryInternalError(t *testing.T) {
	repo := new(mockevmtransactionrepo.Postgres)
	txId := "0xabc"

	repo.On("FindByTxId", mock.Anything, txId).
		Return((*evmtransaction.EvmTransaction)(nil), fmt.Errorf("database connection lost"))

	handler := NewHandler(repo)
	rawResp, err := handler.Execute(context.Background(), Query{
		TxId: txId,
		Dlt:  "EVM",
	})

	assert.Nil(t, rawResp)
	assert.NotNil(t, err)
	repo.AssertExpectations(t)
}

func TestDltIngressGetTransactionQueryHandler_Execute_UnsupportedDlt(t *testing.T) {
	repo := new(mockevmtransactionrepo.Postgres)

	handler := NewHandler(repo)
	rawResp, err := handler.Execute(context.Background(), Query{
		TxId: "0xabc",
		Dlt:  "UNKNOWN",
	})

	assert.Nil(t, rawResp)
	assert.NotNil(t, err)
	repo.AssertNotCalled(t, "FindByTxId", mock.Anything, mock.Anything)
}
