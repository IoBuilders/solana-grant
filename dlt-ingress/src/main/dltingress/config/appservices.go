package dltingressconfig

import (
	"dlt-ingress/src/main/dltingress/internal/app/service/fundaccount"
	"dlt-ingress/src/main/dltingress/internal/app/service/transitfailedtransactiontoretried"
	"dlt-ingress/src/main/dltingress/internal/app/service/txservice"
	"dlt-ingress/src/main/dltingress/internal/infra/contracttransactionbuilder"
	"dlt-ingress/src/main/dltingress/internal/infra/evm"
	"dlt-ingress/src/main/dltingress/internal/infra/nonceprovider"
	"dlt-ingress/src/main/dltingress/internal/infra/svm"
	"dlt-ingress/src/main/dltingress/internal/infra/transactiongasestimator"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
)

type AppServices struct {
	TxService                                   txservice.AppService
	TransitFailedTransactionToRetriedAppService transitfailedtransactiontoretried.AppServiceInterface
	FundAccountAppService                       fundaccount.AppServiceInterface
}

func SetupAppServices(
	nonceProvider nonceprovider.Port,
	blockhashProvider svm.BlockhashProvider,
	transactionBuilderRegistry contracttransactionbuilder.Registry,
	evmClientRegistry evm.ClientRegistry,
	svmClientRegistry svm.ClientRegistry,
	transactionGasEstimationRegistry transactiongasestimator.Registry,
	commandBus command.Bus,
	repositories *DltIngressRepositories,
) *AppServices {
	return &AppServices{
		TxService: *txservice.NewAppService(nonceProvider, blockhashProvider, transactionBuilderRegistry, evmClientRegistry, svmClientRegistry, transactionGasEstimationRegistry),
		TransitFailedTransactionToRetriedAppService: transitfailedtransactiontoretried.NewAppService(commandBus, repositories.FailedTransactionRepo),
		FundAccountAppService:                       fundaccount.NewAppService(repositories.FaucetWalletRepo, repositories.CustodyKeyRepo, evmClientRegistry, svmClientRegistry, commandBus),
	}
}
