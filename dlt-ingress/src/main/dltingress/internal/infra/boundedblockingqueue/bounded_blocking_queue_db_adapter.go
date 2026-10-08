package boundedblockingqueue

import (
	"context"
	"fmt"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/domain/queuelock"
	"dlt-ingress/src/main/dltingress/internal/domain/txqueueslot"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/retry"
)

type BoundedBlockingQueueDbAdapter struct {
	queueLockRepository   queuelock.Repository
	txQueueSlotRepository txqueueslot.Repository
	retryer               retry.Retryer
	retryOptions          retry.Options
	transactionManager    db.TransactionManager
}

func NewBoundedBlockingQueueDbAdapter(
	queueLockRepository queuelock.Repository,
	txQueueSlotRepository txqueueslot.Repository,
	retryer retry.Retryer,
	retryOptions retry.Options,
	transactionManager db.TransactionManager,

) *BoundedBlockingQueueDbAdapter {
	return &BoundedBlockingQueueDbAdapter{
		queueLockRepository:   queueLockRepository,
		txQueueSlotRepository: txQueueSlotRepository,
		retryer:               retryer,
		retryOptions:          retryOptions,
		transactionManager:    transactionManager,
	}
}

func (a *BoundedBlockingQueueDbAdapter) Put(ctx context.Context, networkId string, value uuid.UUID) error {
	networkConfig, err := config.DltIngressConfig.DltIngress.GetDltIngressNetwork(networkId)
	if err != nil {
		return err
	}
	_, err = a.retryer.Execute(ctx, a.retryOptions, func(ctx context.Context) (any, error) {
		// Create new transaction because each retry should be executed in new transaction to avoid having the database lock during all attempts
		retryCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		retryCtx = a.transactionManager.WithNewTransactionValue(retryCtx)
		tx := a.transactionManager.TransactionValue(retryCtx)

		// Lock queue so anyone can count for the same queue
		if _, err := a.queueLockRepository.FindAndLockByNetworkId(retryCtx, networkId); err != nil {
			rollback(retryCtx, tx)
			return nil, err
		}
		// Get count after lock
		count, err := a.txQueueSlotRepository.CountByNetworkIdAndNotExpired(retryCtx, networkId)
		if err != nil {
			rollback(retryCtx, tx)
			return nil, err
		}
		if count >= networkConfig.MaxTxPoolSize {
			// Should retry to find a slot
			rollback(retryCtx, tx)
			return nil, fmt.Errorf("error while trying to put into the queue for network id %s: queue max capacity reached", networkId)
		}
		// Insert new element in the queue
		if err := a.txQueueSlotRepository.Create(retryCtx, txqueueslot.NewTxQueueSlot(value, networkId, config.DltIngressConfig.DltIngress.TxBoundedBlockingQueue.ExpirationTime)); err != nil {
			rollback(retryCtx, tx)
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			rollback(retryCtx, tx)
			return nil, err
		}
		logger.InfoWithCtx(retryCtx, fmt.Sprintf("new value %s put into the queue for network id %s", value, networkId))
		return nil, nil
	})
	return err
}

func (a *BoundedBlockingQueueDbAdapter) Take(ctx context.Context, networkId string) error {
	// Runs the deletion in an independent transaction because it is called on failure paths,
	// where the caller's transaction is about to be rolled back and would undo the release.
	takeCtx := a.transactionManager.WithNewTransactionValue(ctx)
	tx := a.transactionManager.TransactionValue(takeCtx)

	if err := a.txQueueSlotRepository.DeleteFirstByNetworkId(takeCtx, networkId); err != nil {
		rollback(takeCtx, tx)
		return err
	}
	if err := tx.Commit(); err != nil {
		rollback(takeCtx, tx)
		return err
	}
	logger.InfoWithCtx(ctx, fmt.Sprintf("element taken from the queue for network id %s", networkId))
	return nil
}

func rollback(ctx context.Context, tx db.Transaction) {
	if err := tx.Rollback(); err != nil {
		logger.ErrorWithCtx(ctx, "failed to rollback transaction in bounded blocking queue", "error", err)
	}
}
