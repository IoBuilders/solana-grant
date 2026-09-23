package dltingressconfig

import (
	"dlt-ingress/src/main/dltingress/app/command/buildtransaction"
	"dlt-ingress/src/main/dltingress/app/command/createkey"
	"dlt-ingress/src/main/dltingress/app/command/retrytransaction"
	"dlt-ingress/src/main/dltingress/app/command/savefailedtransaction"
	"dlt-ingress/src/main/dltingress/app/command/signandsend"
	"dlt-ingress/src/main/dltingress/app/command/transitfailedtransactiontoretried"
	"dlt-ingress/src/main/dltingress/port/boundedblockingqueue"
	"dlt-ingress/src/main/dltingress/port/custody"
	"dlt-ingress/src/main/dltingress/port/evm"
	"dlt-ingress/src/main/dltingress/port/nonceprovider"
	"dlt-ingress/src/main/dltingress/port/repository"
	"dlt-ingress/src/main/dltingress/port/transactiongasestimator"
	"dlt-ingress/src/main/dltingress/port/transanctionsender"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/app/command/transitfailedeventconsumertopending"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
)

type CommandHandlerDeps struct {
	Repositories                     *repository.DltIngressRepositories
	CustodyProvider                  custody.Port
	EventBus                         event.Bus
	DomainServices                   *DomainServices
	NonceProvider                    nonceprovider.Port
	TransactionSenderRegistry        transanctionsender.Registry
	TransactionGasEstimationRegistry transactiongasestimator.Registry
	BoundedBlockingQueue             boundedblockingqueue.Port
	EvmClientRegistry                evm.ClientRegistry
	AppServices                      *AppServices
	MetricsRegistry                  *metrics.Registry
}

func InitCommandHandlers(deps CommandHandlerDeps) []interface{} {
	return []interface{}{
		createkey.NewCommandHandler(deps.Repositories.CustodyKeyRepo, deps.CustodyProvider, deps.EventBus),
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
		retrytransaction.NewCommandHandler(
			deps.EventBus,
			deps.CustodyProvider,
			deps.NonceProvider,
			deps.TransactionSenderRegistry,
			deps.Repositories.EvmTransactionRepo,
			deps.EvmClientRegistry,
			deps.BoundedBlockingQueue,
			&deps.DomainServices.CustodyKeyExistsService,
		),
		transitfailedeventconsumertopending.NewHandler(deps.EventBus, deps.Repositories.EventConsumerRepo),
		transitfailedtransactiontoretried.NewHandler(deps.EventBus, deps.Repositories.FailedTransactionRepo),
		savefailedtransaction.NewCommandHandler(deps.Repositories.FailedTransactionRepo, deps.EventBus),
	}
}
