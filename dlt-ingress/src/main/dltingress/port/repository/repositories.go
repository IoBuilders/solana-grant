package repository

import (
	"dlt-ingress/src/main/dltingress/port/repository/custodykey"
	"dlt-ingress/src/main/dltingress/port/repository/nonce"
	"dlt-ingress/src/main/dltingress/port/repository/queue"
	"dlt-ingress/src/main/dltingress/port/repository/transaction/evmtransaction"
	"dlt-ingress/src/main/dltingress/port/repository/transaction/failedtransaction"
	"dlt-ingress/src/main/dltingress/port/repository/transaction/svmtransactionrepo"
	"dlt-ingress/src/main/dltingress/port/repository/txqueueslot"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/port/repository/eventstore"
	"gorm.io/gorm"
)

type DltIngressRepositories struct {
	EventConsumerRepo     eventstorerepo.EventConsumerRepository
	EventStoreRepo        eventstorerepo.Repository
	CustodyKeyRepo        custodykeyrepo.Repository
	EvmTransactionRepo    evmtransactionrepo.Repository
	SvmTransactionRepo    svmtransactionrepo.Repository
	QueueLockRepo         queuelockrepo.Repository
	TxQueueSlotRepo       txqueueslotrepo.Repository
	NonceRepo             noncerepo.Repository
	FailedTransactionRepo failedtransactionrepo.Repository
}

func SetupRepositories(gormDB *gorm.DB, tm db.TransactionManager) *DltIngressRepositories {
	return &DltIngressRepositories{
		EventConsumerRepo:     eventstorerepo.NewPostgresEventConsumerRepository(gormDB, tm),
		EventStoreRepo:        eventstorerepo.NewPostgresEventStoreRepository(gormDB, tm),
		CustodyKeyRepo:        custodykeyrepo.NewPostgresCustodyKeyRepository(gormDB, tm),
		EvmTransactionRepo:    evmtransactionrepo.NewPostgresEvmTransactionRepository(gormDB, tm),
		SvmTransactionRepo:    svmtransactionrepo.NewPostgresSVMTransactionRepository(gormDB, tm),
		QueueLockRepo:         queuelockrepo.NewPostgresQueueLockRepository(gormDB, tm),
		TxQueueSlotRepo:       txqueueslotrepo.NewPostgresTxQueueSlotRepository(gormDB, tm),
		NonceRepo:             noncerepo.NewPostgresNonceRepository(gormDB, tm),
		FailedTransactionRepo: failedtransactionrepo.NewPostgresFailedTransactionRepository(gormDB, tm),
	}
}
