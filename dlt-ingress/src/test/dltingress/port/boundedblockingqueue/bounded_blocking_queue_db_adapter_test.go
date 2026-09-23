package boundedblockingqueue

import (
	"context"
	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/port/boundedblockingqueue"
	"dlt-ingress/src/test/dltingress/port/repository"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/retry"
)

func TestDltIngressBoundedBlockingQueueDbAdapter_PutNetworkNotFound(t *testing.T) {
	queueLockRepository := new(mocks.QueueLockRepositoryMock)
	txQueueSlotRepository := new(mocks.TxQueueSlotRepositoryMock)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions()
	transactionManager := new(db.TransactionManagerMock)
	networkId := "testNetworkId"
	config.AppConfig = &config.Config{DltIngress: config.DltIngress{}}
	queue := boundedblockingqueue.NewBoundedBlockingQueueDbAdapter(
		queueLockRepository,
		txQueueSlotRepository,
		retryer,
		retryOptions,
		transactionManager,
	)

	err := queue.Put(context.Background(), networkId, uuid.New())
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("Entity Network with %s not found", networkId), err.Error())
}

func TestDltIngressBoundedBlockingQueueDbAdapter_PutFindAndLockByNetworkIdError(t *testing.T) {
	queueLockRepository := new(mocks.QueueLockRepositoryMock)
	txQueueSlotRepository := new(mocks.TxQueueSlotRepositoryMock)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	expectedErr := fmt.Errorf("persistence test error")
	config.AppConfig = &config.Config{DltIngress: config.DltIngress{Networks: []*config.NetworkConfig{{Id: networkId}}}}
	queue := boundedblockingqueue.NewBoundedBlockingQueueDbAdapter(
		queueLockRepository,
		txQueueSlotRepository,
		retryer,
		retryOptions,
		transactionManager,
	)

	transactionManager.On("WithNewTransactionValue", mock.Anything).Return(context.Background())
	transactionManager.On("TransactionValue", mock.Anything).Return(transaction)
	transaction.On("Rollback").Return(nil)
	queueLockRepository.On("FindAndLockByNetworkId", mock.Anything, networkId).Return(nil, expectedErr)

	err := queue.Put(context.Background(), networkId, uuid.New())
	assert.NotNil(t, err)
	assert.Equal(t, expectedErr.Error(), err.Error())
	transaction.AssertCalled(t, "Rollback")
}

func TestDltIngressBoundedBlockingQueueDbAdapter_PutTxQueueSlotRepositoryCountByNetworkIdAndNotExpiredFindAndLockByNetworkIdError(t *testing.T) {
	queueLockRepository := new(mocks.QueueLockRepositoryMock)
	txQueueSlotRepository := new(mocks.TxQueueSlotRepositoryMock)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	expectedErr := fmt.Errorf("persistence test error")
	config.AppConfig = &config.Config{DltIngress: config.DltIngress{Networks: []*config.NetworkConfig{{Id: networkId}}}}
	queue := boundedblockingqueue.NewBoundedBlockingQueueDbAdapter(
		queueLockRepository,
		txQueueSlotRepository,
		retryer,
		retryOptions,
		transactionManager,
	)

	transactionManager.On("WithNewTransactionValue", mock.Anything).Return(context.Background())
	transactionManager.On("TransactionValue", mock.Anything).Return(transaction)
	transaction.On("Rollback").Return(nil)
	queueLockRepository.On("FindAndLockByNetworkId", mock.Anything, networkId).Return(nil, nil)
	txQueueSlotRepository.On("CountByNetworkIdAndNotExpired", mock.Anything, networkId).Return(0, expectedErr)

	err := queue.Put(context.Background(), networkId, uuid.New())
	assert.NotNil(t, err)
	assert.Equal(t, expectedErr.Error(), err.Error())
	transaction.AssertCalled(t, "Rollback")
}

func TestDltIngressBoundedBlockingQueueDbAdapter_PutMaxCapacityReachError(t *testing.T) {
	queueLockRepository := new(mocks.QueueLockRepositoryMock)
	txQueueSlotRepository := new(mocks.TxQueueSlotRepositoryMock)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	maxTxPoolSize := 100
	config.AppConfig = &config.Config{DltIngress: config.DltIngress{Networks: []*config.NetworkConfig{{Id: networkId, MaxTxPoolSize: maxTxPoolSize}}}}
	queue := boundedblockingqueue.NewBoundedBlockingQueueDbAdapter(
		queueLockRepository,
		txQueueSlotRepository,
		retryer,
		retryOptions,
		transactionManager,
	)

	transactionManager.On("WithNewTransactionValue", mock.Anything).Return(context.Background())
	transactionManager.On("TransactionValue", mock.Anything).Return(transaction)
	transaction.On("Rollback").Return(nil)
	queueLockRepository.On("FindAndLockByNetworkId", mock.Anything, networkId).Return(nil, nil)
	txQueueSlotRepository.On("CountByNetworkIdAndNotExpired", mock.Anything, networkId).Return(maxTxPoolSize, nil)

	err := queue.Put(context.Background(), networkId, uuid.New())
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("error while trying to put into the queue for network id %s: queue max capacity reached", networkId), err.Error())
	transaction.AssertCalled(t, "Rollback")
}

func TestDltIngressBoundedBlockingQueueDbAdapter_PutCreateError(t *testing.T) {
	queueLockRepository := new(mocks.QueueLockRepositoryMock)
	txQueueSlotRepository := new(mocks.TxQueueSlotRepositoryMock)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	maxTxPoolSize := 100
	expectedErr := fmt.Errorf("persistence test error")
	config.AppConfig = &config.Config{DltIngress: config.DltIngress{Networks: []*config.NetworkConfig{{Id: networkId, MaxTxPoolSize: maxTxPoolSize}}}}
	queue := boundedblockingqueue.NewBoundedBlockingQueueDbAdapter(
		queueLockRepository,
		txQueueSlotRepository,
		retryer,
		retryOptions,
		transactionManager,
	)

	transactionManager.On("WithNewTransactionValue", mock.Anything).Return(context.Background())
	transactionManager.On("TransactionValue", mock.Anything).Return(transaction)
	transaction.On("Rollback").Return(nil)
	queueLockRepository.On("FindAndLockByNetworkId", mock.Anything, networkId).Return(nil, nil)
	txQueueSlotRepository.On("CountByNetworkIdAndNotExpired", mock.Anything, networkId).Return(maxTxPoolSize-1, nil)
	txQueueSlotRepository.On("Create", mock.Anything, mock.Anything).Return(expectedErr)

	err := queue.Put(context.Background(), networkId, uuid.New())
	assert.NotNil(t, err)
	assert.Equal(t, expectedErr.Error(), err.Error())
	transaction.AssertCalled(t, "Rollback")
}

func TestDltIngressBoundedBlockingQueueDbAdapter_PutRollbackError(t *testing.T) {
	queueLockRepository := new(mocks.QueueLockRepositoryMock)
	txQueueSlotRepository := new(mocks.TxQueueSlotRepositoryMock)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	maxTxPoolSize := 100
	expectedErr := fmt.Errorf("persistence test error")
	config.AppConfig = &config.Config{DltIngress: config.DltIngress{Networks: []*config.NetworkConfig{{Id: networkId, MaxTxPoolSize: maxTxPoolSize}}}}
	queue := boundedblockingqueue.NewBoundedBlockingQueueDbAdapter(
		queueLockRepository,
		txQueueSlotRepository,
		retryer,
		retryOptions,
		transactionManager,
	)

	transactionManager.On("WithNewTransactionValue", mock.Anything).Return(context.Background())
	transactionManager.On("TransactionValue", mock.Anything).Return(transaction)
	transaction.On("Rollback").Return(fmt.Errorf("rollback error"))
	queueLockRepository.On("FindAndLockByNetworkId", mock.Anything, networkId).Return(nil, nil)
	txQueueSlotRepository.On("CountByNetworkIdAndNotExpired", mock.Anything, networkId).Return(maxTxPoolSize-1, nil)
	txQueueSlotRepository.On("Create", mock.Anything, mock.Anything).Return(expectedErr)

	err := queue.Put(context.Background(), networkId, uuid.New())
	assert.NotNil(t, err)
	assert.Equal(t, expectedErr.Error(), err.Error())
	transaction.AssertCalled(t, "Rollback")
}

func TestDltIngressBoundedBlockingQueueDbAdapter_PutCommitError(t *testing.T) {
	queueLockRepository := new(mocks.QueueLockRepositoryMock)
	txQueueSlotRepository := new(mocks.TxQueueSlotRepositoryMock)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	maxTxPoolSize := 100
	expectedErr := fmt.Errorf("persistence test error")
	config.AppConfig = &config.Config{DltIngress: config.DltIngress{Networks: []*config.NetworkConfig{{Id: networkId, MaxTxPoolSize: maxTxPoolSize}}}}
	queue := boundedblockingqueue.NewBoundedBlockingQueueDbAdapter(
		queueLockRepository,
		txQueueSlotRepository,
		retryer,
		retryOptions,
		transactionManager,
	)

	transactionManager.On("WithNewTransactionValue", mock.Anything).Return(context.Background())
	transactionManager.On("TransactionValue", mock.Anything).Return(transaction)
	queueLockRepository.On("FindAndLockByNetworkId", mock.Anything, networkId).Return(nil, nil)
	txQueueSlotRepository.On("CountByNetworkIdAndNotExpired", mock.Anything, networkId).Return(maxTxPoolSize-1, nil)
	txQueueSlotRepository.On("Create", mock.Anything, mock.Anything).Return(nil)
	transaction.On("Rollback").Return(nil)
	transaction.On("Commit").Return(expectedErr)

	err := queue.Put(context.Background(), networkId, uuid.New())
	assert.NotNil(t, err)
	assert.Equal(t, expectedErr.Error(), err.Error())
	transaction.AssertCalled(t, "Rollback")
}

func TestDltIngressBoundedBlockingQueueDbAdapter_PutOk(t *testing.T) {
	queueLockRepository := new(mocks.QueueLockRepositoryMock)
	txQueueSlotRepository := new(mocks.TxQueueSlotRepositoryMock)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	maxTxPoolSize := 100
	config.AppConfig = &config.Config{DltIngress: config.DltIngress{Networks: []*config.NetworkConfig{{Id: networkId, MaxTxPoolSize: maxTxPoolSize}}}}
	queue := boundedblockingqueue.NewBoundedBlockingQueueDbAdapter(
		queueLockRepository,
		txQueueSlotRepository,
		retryer,
		retryOptions,
		transactionManager,
	)

	transactionManager.On("WithNewTransactionValue", mock.Anything).Return(context.Background())
	transactionManager.On("TransactionValue", mock.Anything).Return(transaction)
	queueLockRepository.On("FindAndLockByNetworkId", mock.Anything, networkId).Return(nil, nil)
	txQueueSlotRepository.On("CountByNetworkIdAndNotExpired", mock.Anything, networkId).Return(maxTxPoolSize-10, nil)
	txQueueSlotRepository.On("Create", mock.Anything, mock.Anything).Return(nil)
	transaction.On("Rollback").Return(nil)
	transaction.On("Commit").Return(nil)

	err := queue.Put(context.Background(), networkId, uuid.New())
	assert.Nil(t, err)
	transaction.AssertNotCalled(t, "Rollback")
	transaction.AssertCalled(t, "Commit")
}

func TestDltIngressBoundedBlockingQueueDbAdapter_TakeError(t *testing.T) {
	queueLockRepository := new(mocks.QueueLockRepositoryMock)
	txQueueSlotRepository := new(mocks.TxQueueSlotRepositoryMock)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	networkId := "testNetworkId"
	expectedErr := fmt.Errorf("persistence test error")
	queue := boundedblockingqueue.NewBoundedBlockingQueueDbAdapter(
		queueLockRepository,
		txQueueSlotRepository,
		retryer,
		retryOptions,
		transactionManager,
	)

	txQueueSlotRepository.On("DeleteFirstByNetworkId", mock.Anything, mock.Anything).Return(expectedErr)

	err := queue.Take(context.Background(), networkId)
	assert.NotNil(t, err)
	assert.Equal(t, expectedErr.Error(), err.Error())
}

func TestDltIngressBoundedBlockingQueueDbAdapter_TakeOk(t *testing.T) {
	queueLockRepository := new(mocks.QueueLockRepositoryMock)
	txQueueSlotRepository := new(mocks.TxQueueSlotRepositoryMock)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	networkId := "testNetworkId"
	queue := boundedblockingqueue.NewBoundedBlockingQueueDbAdapter(
		queueLockRepository,
		txQueueSlotRepository,
		retryer,
		retryOptions,
		transactionManager,
	)

	txQueueSlotRepository.On("DeleteFirstByNetworkId", mock.Anything, mock.Anything).Return(nil)

	err := queue.Take(context.Background(), networkId)
	assert.Nil(t, err)
}
