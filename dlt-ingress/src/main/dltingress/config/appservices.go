package dltingressconfig

import (
	"dlt-ingress/src/main/dltingress/app/service/transitfailedtransactiontoretried"
	"dlt-ingress/src/main/dltingress/app/service/txservice"
	"dlt-ingress/src/main/dltingress/port/contracttransactionbuilder"
	"dlt-ingress/src/main/dltingress/port/evm"
	"dlt-ingress/src/main/dltingress/port/nonceprovider"
	"dlt-ingress/src/main/dltingress/port/repository"
	"dlt-ingress/src/main/dltingress/port/svm"
	"dlt-ingress/src/main/dltingress/port/transactiongasestimator"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
)

type AppServices struct {
	TxService                                   txservice.AppService
	TransitFailedTransactionToRetriedAppService transitfailedtransactiontoretried.AppServiceInterface
}

func SetupAppServices(
	nonceProvider nonceprovider.Port,
	blockhashProvider svm.BlockhashProvider,
	transactionBuilderRegistry contracttransactionbuilder.Registry,
	evmClientRegistry evm.ClientRegistry,
	svmClientRegistry svm.ClientRegistry,
	transactionGasEstimationRegistry transactiongasestimator.Registry,
	commandBus command.Bus,
	repositories *repository.DltIngressRepositories,
) *AppServices {
	return &AppServices{
		TxService: *txservice.NewAppService(nonceProvider, blockhashProvider, transactionBuilderRegistry, evmClientRegistry, svmClientRegistry, transactionGasEstimationRegistry),
		TransitFailedTransactionToRetriedAppService: transitfailedtransactiontoretried.NewAppService(commandBus, repositories.FailedTransactionRepo),
	}
}
