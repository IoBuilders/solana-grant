package dltingressconfig

import (
	"dlt-ingress/src/main/dltingress/internal/app/command/buildtransaction"
	"dlt-ingress/src/main/dltingress/internal/app/command/createfaucetwallet"
	"dlt-ingress/src/main/dltingress/internal/app/command/createkey"
	"dlt-ingress/src/main/dltingress/internal/app/command/retryevmtransaction"
	"dlt-ingress/src/main/dltingress/internal/app/command/retrysvmtransaction"
	"dlt-ingress/src/main/dltingress/internal/app/command/savefailedtransaction"
	"dlt-ingress/src/main/dltingress/internal/app/command/signandsend"
	"dlt-ingress/src/main/dltingress/internal/app/command/transitfailedtransactiontoretried"
	"dlt-ingress/src/main/dltingress/internal/app/command/updatefaucetwallet"
	"dlt-ingress/src/main/dltingress/internal/infra/boundedblockingqueue"
	"dlt-ingress/src/main/dltingress/internal/infra/custody"
	"dlt-ingress/src/main/dltingress/internal/infra/evm"
	"dlt-ingress/src/main/dltingress/internal/infra/nonceprovider"
	"dlt-ingress/src/main/dltingress/internal/infra/svm"
	"dlt-ingress/src/main/dltingress/internal/infra/transactiongasestimator"
	"dlt-ingress/src/main/dltingress/internal/infra/transanctionsender"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/app/command/transitfailedeventconsumertopending"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
)

type CommandHandlerDeps struct {
	Repositories                     *DltIngressRepositories
	CustodyProvider                  custody.Port
	EventBus                         event.Bus
	DomainServices                   *DomainServices
	NonceProvider                    nonceprovider.Port
	TransactionSenderRegistry        transanctionsender.Registry
	TransactionGasEstimationRegistry transactiongasestimator.Registry
	BoundedBlockingQueue             boundedblockingqueue.Port
	EvmClientRegistry                evm.ClientRegistry
	SvmClientRegistry                svm.ClientRegistry
	AppServices                      *AppServices
	MetricsRegistry                  *metrics.Registry
}

func InitCommandHandlers(deps CommandHandlerDeps) []interface{} {
	return []interface{}{
		createkey.NewCommandHandler(deps.Repositories.CustodyKeyRepo, deps.CustodyProvider, deps.EventBus),
		createfaucetwallet.NewCommandHandler(deps.Repositories.FaucetWalletRepo, deps.Repositories.CustodyKeyRepo, deps.CustodyProvider, deps.EventBus),
		updatefaucetwallet.NewCommandHandler(deps.Repositories.FaucetWalletRepo, deps.EventBus),
		signandsend.NewCommandHandlerMetrics(deps.MetricsRegistry, signandsend.NewCommandHandler(
			deps.EventBus,
			&deps.AppServices.TxService,
			deps.NonceProvider,
			deps.CustodyProvider,
			deps.TransactionSenderRegistry,
			&deps.DomainServices.CustodyKeyExistMultipleService,
			deps.BoundedBlockingQueue,
			deps.Repositories.EvmTransactionRepo,
			deps.Repositories.SvmTransactionRepo,
		)),
		buildtransaction.NewCommandHandler(&deps.AppServices.TxService),
		retryevmtransaction.NewCommandHandler(
			deps.EventBus,
			deps.CustodyProvider,
			deps.NonceProvider,
			deps.TransactionSenderRegistry,
			deps.Repositories.EvmTransactionRepo,
			deps.EvmClientRegistry,
			deps.BoundedBlockingQueue,
			&deps.DomainServices.CustodyKeyExistsService,
		),
		retrysvmtransaction.NewCommandHandler(
			deps.EventBus,
			deps.CustodyProvider,
			deps.SvmClientRegistry,
			deps.TransactionSenderRegistry,
			deps.Repositories.SvmTransactionRepo,
			deps.BoundedBlockingQueue,
			&deps.DomainServices.CustodyKeyExistMultipleService,
		),
		transitfailedeventconsumertopending.NewHandler(deps.EventBus, deps.Repositories.EventConsumerRepo),
		transitfailedtransactiontoretried.NewHandler(deps.EventBus, deps.Repositories.FailedTransactionRepo),
		savefailedtransaction.NewCommandHandler(deps.Repositories.FailedTransactionRepo, deps.EventBus),
	}
}
