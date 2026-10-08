package dltingressconfig

import (
	"dlt-ingress/src/main/dltingress/internal/app/query/contractread"
	"dlt-ingress/src/main/dltingress/internal/app/query/getcustodykey"
	"dlt-ingress/src/main/dltingress/internal/app/query/getfailedtransactions"
	"dlt-ingress/src/main/dltingress/internal/app/query/getfaucetwallet"
	"dlt-ingress/src/main/dltingress/internal/app/query/gettransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/contractcallbuilder"
	"dlt-ingress/src/main/dltingress/internal/infra/contractcaller"
	"dlt-ingress/src/main/dltingress/internal/infra/evm"
	"dlt-ingress/src/main/dltingress/internal/infra/svm"
)

type QueryHandlerDeps struct {
	Repositories                *DltIngressRepositories
	ContractCallBuilderRegistry contractcallbuilder.Registry
	ContractCallerRegistry      contractcaller.Registry
	EvmClientRegistry           evm.ClientRegistry
	SvmClientRegistry           svm.ClientRegistry
}

func SetupQueryHandlers(deps QueryHandlerDeps) []any {
	return []any{
		getcustodykey.NewHandler(deps.Repositories.CustodyKeyRepo),
		gettransaction.NewHandler(deps.Repositories.EvmTransactionRepo),
		getfailedtransactions.NewHandler(deps.Repositories.FailedTransactionRepo),
		contractread.NewHandler(deps.ContractCallBuilderRegistry, deps.ContractCallerRegistry),
		getfaucetwallet.NewHandler(deps.Repositories.FaucetWalletRepo, deps.Repositories.CustodyKeyRepo, deps.EvmClientRegistry, deps.SvmClientRegistry),
	}
}
