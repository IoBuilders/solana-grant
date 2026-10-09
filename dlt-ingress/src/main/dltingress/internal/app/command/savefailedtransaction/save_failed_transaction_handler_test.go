package savefailedtransaction

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/transaction/failedtransaction/mock"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	testbus "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/event"
)

func newHandler() (*CommandHandler, *mockfailedtransactionrepo.Postgres, *testbus.EventBusMock) {
	repo := new(mockfailedtransactionrepo.Postgres)
	eventBus := new(testbus.EventBusMock)
	return NewCommandHandler(repo, eventBus), repo, eventBus
}

func TestDltIngressSaveFailedTransactionCommandHandler_Handle_Success(t *testing.T) {
	handler, repo, eventBus := newHandler()

	repo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
	eventBus.On("Publish", mock.Anything, mock.AnythingOfType("failedtransaction.SavedEvent")).Return(nil).Once()

	cmd := &Command{TxId: "testTxId", NetworkId: "testNetworkId", ErrorDetails: "test error"}
	resp, err := handler.Handle(context.Background(), cmd)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, cmd.TxId, resp.TxId)
	assert.Equal(t, cmd.NetworkId, resp.NetworkId)
	assert.Equal(t, cmd.ErrorDetails, resp.ErrorDetails)
	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestDltIngressSaveFailedTransactionCommandHandler_Handle_Repo_Error(t *testing.T) {
	handler, repo, eventBus := newHandler()

	repo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError).Once()

	cmd := &Command{TxId: "testTxId", NetworkId: "testNetworkId", ErrorDetails: "test error"}
	resp, err := handler.Handle(context.Background(), cmd)

	assert.ErrorIs(t, err, assert.AnError)
	assert.Nil(t, resp)
	repo.AssertExpectations(t)
	eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestDltIngressSaveFailedTransactionCommandHandler_Handle_Event_Bus_Error(t *testing.T) {
	handler, repo, eventBus := newHandler()

	repo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
	eventBus.On("Publish", mock.Anything, mock.AnythingOfType("failedtransaction.SavedEvent")).Return(assert.AnError).Once()

	cmd := &Command{TxId: "testTxId", NetworkId: "testNetworkId", ErrorDetails: "test error"}
	resp, err := handler.Handle(context.Background(), cmd)

	assert.ErrorIs(t, err, assert.AnError)
	assert.Nil(t, resp)
	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}
