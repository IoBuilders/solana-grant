package boundedblockingqueue

import (
	"context"
	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/queuelock/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/txqueueslot/mock"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/retry"
)

func TestDltIngressBoundedBlockingQueueDbAdapter_PutNetworkNotFound(t *testing.T) {
	queueLockRepository := new(mockqueuelockrepo.Postgres)
	txQueueSlotRepository := new(mocktxqueueslotrepo.Postgres)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions()
	transactionManager := new(db.TransactionManagerMock)
	networkId := "testNetworkId"
	config.DltIngressConfig = &config.Config{DltIngress: config.DltIngress{}}
	queue := NewBoundedBlockingQueueDbAdapter(
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
	queueLockRepository := new(mockqueuelockrepo.Postgres)
	txQueueSlotRepository := new(mocktxqueueslotrepo.Postgres)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	expectedErr := fmt.Errorf("persistence test error")
	config.DltIngressConfig = &config.Config{DltIngress: config.DltIngress{Networks: []*config.NetworkConfig{{Id: networkId}}}}
	queue := NewBoundedBlockingQueueDbAdapter(
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
	queueLockRepository := new(mockqueuelockrepo.Postgres)
	txQueueSlotRepository := new(mocktxqueueslotrepo.Postgres)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	expectedErr := fmt.Errorf("persistence test error")
	config.DltIngressConfig = &config.Config{DltIngress: config.DltIngress{Networks: []*config.NetworkConfig{{Id: networkId}}}}
	queue := NewBoundedBlockingQueueDbAdapter(
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
	queueLockRepository := new(mockqueuelockrepo.Postgres)
	txQueueSlotRepository := new(mocktxqueueslotrepo.Postgres)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	maxTxPoolSize := 100
	config.DltIngressConfig = &config.Config{DltIngress: config.DltIngress{Networks: []*config.NetworkConfig{{Id: networkId, MaxTxPoolSize: maxTxPoolSize}}}}
	queue := NewBoundedBlockingQueueDbAdapter(
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
	queueLockRepository := new(mockqueuelockrepo.Postgres)
	txQueueSlotRepository := new(mocktxqueueslotrepo.Postgres)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	maxTxPoolSize := 100
	expectedErr := fmt.Errorf("persistence test error")
	config.DltIngressConfig = &config.Config{DltIngress: config.DltIngress{Networks: []*config.NetworkConfig{{Id: networkId, MaxTxPoolSize: maxTxPoolSize}}}}
	queue := NewBoundedBlockingQueueDbAdapter(
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
	queueLockRepository := new(mockqueuelockrepo.Postgres)
	txQueueSlotRepository := new(mocktxqueueslotrepo.Postgres)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	maxTxPoolSize := 100
	expectedErr := fmt.Errorf("persistence test error")
	config.DltIngressConfig = &config.Config{DltIngress: config.DltIngress{Networks: []*config.NetworkConfig{{Id: networkId, MaxTxPoolSize: maxTxPoolSize}}}}
	queue := NewBoundedBlockingQueueDbAdapter(
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
	queueLockRepository := new(mockqueuelockrepo.Postgres)
	txQueueSlotRepository := new(mocktxqueueslotrepo.Postgres)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	maxTxPoolSize := 100
	expectedErr := fmt.Errorf("persistence test error")
	config.DltIngressConfig = &config.Config{DltIngress: config.DltIngress{Networks: []*config.NetworkConfig{{Id: networkId, MaxTxPoolSize: maxTxPoolSize}}}}
	queue := NewBoundedBlockingQueueDbAdapter(
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
	queueLockRepository := new(mockqueuelockrepo.Postgres)
	txQueueSlotRepository := new(mocktxqueueslotrepo.Postgres)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	maxTxPoolSize := 100
	config.DltIngressConfig = &config.Config{DltIngress: config.DltIngress{Networks: []*config.NetworkConfig{{Id: networkId, MaxTxPoolSize: maxTxPoolSize}}}}
	queue := NewBoundedBlockingQueueDbAdapter(
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

func TestDltIngressBoundedBlockingQueueDbAdapter_TakeDeleteError(t *testing.T) {
	queueLockRepository := new(mockqueuelockrepo.Postgres)
	txQueueSlotRepository := new(mocktxqueueslotrepo.Postgres)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	expectedErr := fmt.Errorf("persistence test error")
	queue := NewBoundedBlockingQueueDbAdapter(
		queueLockRepository,
		txQueueSlotRepository,
		retryer,
		retryOptions,
		transactionManager,
	)

	transactionManager.On("WithNewTransactionValue", mock.Anything).Return(context.Background())
	transactionManager.On("TransactionValue", mock.Anything).Return(transaction)
	transaction.On("Rollback").Return(nil)
	txQueueSlotRepository.On("DeleteFirstByNetworkId", mock.Anything, networkId).Return(expectedErr)

	err := queue.Take(context.Background(), networkId)
	assert.NotNil(t, err)
	assert.Equal(t, expectedErr.Error(), err.Error())
	transaction.AssertCalled(t, "Rollback")
	transaction.AssertNotCalled(t, "Commit")
}

func TestDltIngressBoundedBlockingQueueDbAdapter_TakeCommitError(t *testing.T) {
	queueLockRepository := new(mockqueuelockrepo.Postgres)
	txQueueSlotRepository := new(mocktxqueueslotrepo.Postgres)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	expectedErr := fmt.Errorf("persistence test error")
	queue := NewBoundedBlockingQueueDbAdapter(
		queueLockRepository,
		txQueueSlotRepository,
		retryer,
		retryOptions,
		transactionManager,
	)

	transactionManager.On("WithNewTransactionValue", mock.Anything).Return(context.Background())
	transactionManager.On("TransactionValue", mock.Anything).Return(transaction)
	transaction.On("Rollback").Return(nil)
	transaction.On("Commit").Return(expectedErr)
	txQueueSlotRepository.On("DeleteFirstByNetworkId", mock.Anything, networkId).Return(nil)

	err := queue.Take(context.Background(), networkId)
	assert.NotNil(t, err)
	assert.Equal(t, expectedErr.Error(), err.Error())
	transaction.AssertCalled(t, "Rollback")
}

func TestDltIngressBoundedBlockingQueueDbAdapter_TakeOk(t *testing.T) {
	queueLockRepository := new(mockqueuelockrepo.Postgres)
	txQueueSlotRepository := new(mocktxqueueslotrepo.Postgres)
	retryer := retry.NewCustomRetryer()
	retryOptions := retry.NewOptions(retry.WithMaxAttempts(1))
	transactionManager := new(db.TransactionManagerMock)
	transaction := new(db.TransactionMock)
	networkId := "testNetworkId"
	queue := NewBoundedBlockingQueueDbAdapter(
		queueLockRepository,
		txQueueSlotRepository,
		retryer,
		retryOptions,
		transactionManager,
	)

	transactionManager.On("WithNewTransactionValue", mock.Anything).Return(context.Background())
	transactionManager.On("TransactionValue", mock.Anything).Return(transaction)
	transaction.On("Commit").Return(nil)
	txQueueSlotRepository.On("DeleteFirstByNetworkId", mock.Anything, networkId).Return(nil)

	err := queue.Take(context.Background(), networkId)
	assert.Nil(t, err)
	txQueueSlotRepository.AssertCalled(t, "DeleteFirstByNetworkId", mock.Anything, networkId)
	transaction.AssertCalled(t, "Commit")
	transaction.AssertNotCalled(t, "Rollback")
}
