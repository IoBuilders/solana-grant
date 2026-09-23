package dltingressconfig

import (
	"dlt-ingress/src/main/dltingress/app/query/contractread"
	"dlt-ingress/src/main/dltingress/app/query/getcustodykey"
	"dlt-ingress/src/main/dltingress/app/query/getfailedtransactions"
	"dlt-ingress/src/main/dltingress/app/query/gettransaction"
	"dlt-ingress/src/main/dltingress/port/contractcallbuilder"
	"dlt-ingress/src/main/dltingress/port/contractcaller"
	"dlt-ingress/src/main/dltingress/port/repository"
)

type QueryHandlerDeps struct {
	Repositories                *repository.DltIngressRepositories
	ContractCallBuilderRegistry contractcallbuilder.Registry
	ContractCallerRegistry      contractcaller.Registry
}

func SetupQueryHandlers(deps QueryHandlerDeps) []any {
	return []any{
		getcustodykey.NewHandler(deps.Repositories.CustodyKeyRepo),
		gettransaction.NewHandler(deps.Repositories.EvmTransactionRepo),
		getfailedtransactions.NewHandler(deps.Repositories.FailedTransactionRepo),
		contractread.NewHandler(deps.ContractCallBuilderRegistry, deps.ContractCallerRegistry),
	}
}
