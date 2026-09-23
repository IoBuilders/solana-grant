package transitfailedtransactiontoretried

import (
	"context"
	"dlt-ingress/src/main/dltingress/app/command/transitfailedtransactiontoretried"
	"dlt-ingress/src/main/dltingress/domain/transaction/failedtransaction"
	mocks "dlt-ingress/src/test/dltingress/port/repository"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/basemodel"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/event"
)

func buildFailedTransaction(status failedtransaction.Status) *failedtransaction.FailedTransaction {
	return &failedtransaction.FailedTransaction{
		Model:  basemodel.Model{Id: uuid.New()},
		TxId:   "test transaction id",
		Status: status,
	}
}

func TestDltIngressTransitFailedTransactionToRetriedHandler_Handle_Success(t *testing.T) {
	eventBus := new(testevent.EventBusMock)
	repo := new(mocks.FailedTransactionRepositoryMock)
	handler := transitfailedtransactiontoretried.NewHandler(eventBus, repo)

	failedTransaction := buildFailedTransaction(failedtransaction.StatusNotRetried)
	repo.On("FindByTxId", mock.Anything, failedTransaction.TxId).Return(failedTransaction, nil)
	repo.On("Save", mock.Anything, mock.Anything).Return(nil)
	eventBus.On("Publish", mock.Anything, mock.MatchedBy(func(evt failedtransaction.TransitFailedTransactionToRetriedEvent) bool {
		return evt.TxId == failedTransaction.TxId && evt.NetworkId == failedTransaction.NetworkId
	})).Return(nil).Once()

	cmd := transitfailedtransactiontoretried.Command{TxId: failedTransaction.TxId}

	resp, err := handler.Handle(context.Background(), &cmd)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, failedTransaction.TxId, resp.TxId)
	assert.Equal(t, failedTransaction.NetworkId, resp.NetworkId)
	assert.Equal(t, failedtransaction.StatusRetried, failedTransaction.Status)
	repo.AssertExpectations(t)
}

func TestDltIngressTransitFailedTransactionToRetriedHandler_Handle_NotificationNotFoundError(t *testing.T) {
	eventBus := new(testevent.EventBusMock)
	repo := new(mocks.FailedTransactionRepositoryMock)
	handler := transitfailedtransactiontoretried.NewHandler(eventBus, repo)

	failedTransactionTxId := "test transaction id"
	repo.On("FindByTxId", mock.Anything, failedTransactionTxId).
		Return(nil, domainerrors.NewEntityNotFoundDomainError("FailedTransaction", failedTransactionTxId))

	cmd := transitfailedtransactiontoretried.Command{TxId: failedTransactionTxId}

	_, err := handler.Handle(context.Background(), &cmd)

	assert.NotNil(t, err)
	assert.Equal(t, coreerror.ErrorCode("FAILED_TRANSACTION_NOT_FOUND"), err.(coreerror.DomainError).ErrorCode())
	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestDltIngressTransitFailedTransactionToRetriedHandler_Handle_InvalidStatusError(t *testing.T) {
	eventBus := new(testevent.EventBusMock)
	repo := new(mocks.FailedTransactionRepositoryMock)
	handler := transitfailedtransactiontoretried.NewHandler(eventBus, repo)

	for _, status := range []failedtransaction.Status{failedtransaction.StatusRetried} {
		failedTransaction := buildFailedTransaction(status)
		repo.On("FindByTxId", mock.Anything, failedTransaction.TxId).Return(failedTransaction, nil)

		cmd := transitfailedtransactiontoretried.Command{TxId: failedTransaction.TxId}

		_, err := handler.Handle(context.Background(), &cmd)

		assert.NotNil(t, err)
		assert.Equal(t, coreerror.ErrorCode("FAILED_TRANSACTION_UNEXPECTED_STATUS"), err.(coreerror.DomainError).ErrorCode())
		repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	}
}

func TestDltIngressTransitFailedTransactionToRetriedHandler_Handle_DatabaseError(t *testing.T) {
	eventBus := new(testevent.EventBusMock)
	repo := new(mocks.FailedTransactionRepositoryMock)
	handler := transitfailedtransactiontoretried.NewHandler(eventBus, repo)

	customErr := errors.New("database error")
	failedTransaction := buildFailedTransaction(failedtransaction.StatusNotRetried)
	repo.On("FindByTxId", mock.Anything, failedTransaction.TxId).Return(failedTransaction, nil)
	repo.On("Save", mock.Anything, mock.Anything).Return(customErr)

	cmd := transitfailedtransactiontoretried.Command{TxId: failedTransaction.TxId}

	_, err := handler.Handle(context.Background(), &cmd)

	assert.NotNil(t, err)
	assert.ErrorIs(t, err, customErr)
	repo.AssertExpectations(t)
}

func TestDltIngressTransitFailedTransactionToRetriedHandler_Handle_EventPublishError(t *testing.T) {
	eventBus := new(testevent.EventBusMock)
	repo := new(mocks.FailedTransactionRepositoryMock)
	handler := transitfailedtransactiontoretried.NewHandler(eventBus, repo)

	customErr := errors.New("publish error")
	failedTransaction := buildFailedTransaction(failedtransaction.StatusNotRetried)
	repo.On("FindByTxId", mock.Anything, failedTransaction.TxId).Return(failedTransaction, nil)
	repo.On("Save", mock.Anything, mock.Anything).Return(nil)
	eventBus.On("Publish", mock.Anything, mock.Anything).Return(customErr).Once()

	cmd := transitfailedtransactiontoretried.Command{TxId: failedTransaction.TxId}

	_, err := handler.Handle(context.Background(), &cmd)

	assert.NotNil(t, err)
	assert.ErrorIs(t, err, customErr)
	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}
