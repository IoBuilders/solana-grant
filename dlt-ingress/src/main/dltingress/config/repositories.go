package dltingressconfig

import (
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	"dlt-ingress/src/main/dltingress/internal/domain/faucetwallet"
	"dlt-ingress/src/main/dltingress/internal/domain/nonce"
	"dlt-ingress/src/main/dltingress/internal/domain/queuelock"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/evmtransaction"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/failedtransaction"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/svmtransaction"
	"dlt-ingress/src/main/dltingress/internal/domain/txqueueslot"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/custodykey"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/faucetwallet"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/nonce"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/queuelock"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/transaction/evmtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/transaction/failedtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/transaction/svmtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/txqueueslot"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/port/repository/eventstore"
	"gorm.io/gorm"
)

type DltIngressRepositories struct {
	EventConsumerRepo     eventstorerepo.EventConsumerRepository
	EventStoreRepo        eventstorerepo.Repository
	CustodyKeyRepo        custodykey.Repository
	FaucetWalletRepo      faucetwallet.Repository
	EvmTransactionRepo    evmtransaction.Repository
	SvmTransactionRepo    svmtransaction.Repository
	QueueLockRepo         queuelock.Repository
	TxQueueSlotRepo       txqueueslot.Repository
	NonceRepo             nonce.Repository
	FailedTransactionRepo failedtransaction.Repository
}

func SetupRepositories(gormDB *gorm.DB, tm db.TransactionManager) *DltIngressRepositories {
	return &DltIngressRepositories{
		EventConsumerRepo:     eventstorerepo.NewPostgresEventConsumerRepository(gormDB, tm),
		EventStoreRepo:        eventstorerepo.NewPostgresEventStoreRepository(gormDB, tm),
		CustodyKeyRepo:        custodykeyrepo.NewPostgres(gormDB, tm),
		FaucetWalletRepo:      faucetwalletrepo.NewPostgres(gormDB, tm),
		EvmTransactionRepo:    evmtransactionrepo.NewPostgres(gormDB, tm),
		SvmTransactionRepo:    svmtransactionrepo.NewPostgres(gormDB, tm),
		QueueLockRepo:         queuelockrepo.NewPostgres(gormDB, tm),
		TxQueueSlotRepo:       txqueueslotrepo.NewPostgres(gormDB, tm),
		NonceRepo:             noncerepo.NewPostgres(gormDB, tm),
		FailedTransactionRepo: failedtransactionrepo.NewPostgresFailedTransactionRepository(gormDB, tm),
	}
}
