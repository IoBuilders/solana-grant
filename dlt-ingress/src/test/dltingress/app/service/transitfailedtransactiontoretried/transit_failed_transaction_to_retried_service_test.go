package transitfailedtransactiontoretried

import (
	"context"
	transitfailedtransactiontoretriedcommand "dlt-ingress/src/main/dltingress/app/command/transitfailedtransactiontoretried"
	"dlt-ingress/src/main/dltingress/app/service/transitfailedtransactiontoretried"
	"dlt-ingress/src/main/dltingress/domain/transaction/failedtransaction"
	"dlt-ingress/src/test/dltingress/port/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/command"
)

func TestDltIngressTransitFailedTransactionToRetriedAppService_Success(t *testing.T) {
	commandBus := new(commandbusmock.CommandBusMock)
	repo := new(mocks.FailedTransactionRepositoryMock)
	appService := transitfailedtransactiontoretried.NewAppService(commandBus, repo)

	txId := "test transaction id"
	repo.On("ExistByTxIdAndStatus", mock.Anything, txId, failedtransaction.StatusNotRetried).Return(true, nil).Once()
	commandBus.On("Dispatch", mock.Anything, mock.MatchedBy(func(cmd *transitfailedtransactiontoretriedcommand.Command) bool {
		return cmd.TxId == txId
	})).Return(nil, nil).Once()

	err := appService.Execute(context.Background(), &transitfailedtransactiontoretried.Request{TxId: txId})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	commandBus.AssertExpectations(t)
}

func TestDltIngressTransitFailedTransactionToRetriedAppService_Failed_Transaction_Not_Exist(t *testing.T) {
	commandBus := new(commandbusmock.CommandBusMock)
	repo := new(mocks.FailedTransactionRepositoryMock)
	appService := transitfailedtransactiontoretried.NewAppService(commandBus, repo)

	txId := "test transaction id"
	repo.On("ExistByTxIdAndStatus", mock.Anything, txId, failedtransaction.StatusNotRetried).Return(false, nil).Once()

	err := appService.Execute(context.Background(), &transitfailedtransactiontoretried.Request{TxId: txId})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	commandBus.AssertNotCalled(t, "Dispatch", mock.Anything, mock.Anything)
}

func TestDltIngressTransitFailedTransactionToRetriedAppService_Repository_Error(t *testing.T) {
	commandBus := new(commandbusmock.CommandBusMock)
	repo := new(mocks.FailedTransactionRepositoryMock)
	appService := transitfailedtransactiontoretried.NewAppService(commandBus, repo)

	txId := "test transaction id"
	repo.On("ExistByTxIdAndStatus", mock.Anything, txId, failedtransaction.StatusNotRetried).Return(false, assert.AnError).Once()

	err := appService.Execute(context.Background(), &transitfailedtransactiontoretried.Request{TxId: txId})

	assert.ErrorIs(t, err, assert.AnError)
	repo.AssertExpectations(t)
	commandBus.AssertNotCalled(t, "Dispatch", mock.Anything, mock.Anything)
}

func TestDltIngressTransitFailedTransactionToRetriedAppService_Command_Bus_Error(t *testing.T) {
	commandBus := new(commandbusmock.CommandBusMock)
	repo := new(mocks.FailedTransactionRepositoryMock)
	appService := transitfailedtransactiontoretried.NewAppService(commandBus, repo)

	txId := "test transaction id"
	repo.On("ExistByTxIdAndStatus", mock.Anything, txId, failedtransaction.StatusNotRetried).Return(true, nil).Once()
	commandBus.On("Dispatch", mock.Anything, mock.MatchedBy(func(cmd *transitfailedtransactiontoretriedcommand.Command) bool {
		return cmd.TxId == txId
	})).Return(nil, assert.AnError).Once()

	err := appService.Execute(context.Background(), &transitfailedtransactiontoretried.Request{TxId: txId})

	assert.ErrorIs(t, err, assert.AnError)
	repo.AssertExpectations(t)
	commandBus.AssertExpectations(t)
}
