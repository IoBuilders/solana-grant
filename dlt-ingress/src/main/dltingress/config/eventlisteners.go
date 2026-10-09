package dltingressconfig

import (
	"dlt-ingress/src/main/core/shared"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction"
	"dlt-ingress/src/main/dltingress/internal/infra/listener/custodykey/keycreated"
	"dlt-ingress/src/main/dltingress/internal/infra/listener/failedtransaction/transitfailedtransactiontoretried"
	"dlt-ingress/src/main/dltingress/internal/infra/listener/fundaccount"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/config"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/retry"
)

func RegisterEvents(registry *event.ListenerRegistry) {
	registry.RegisterPrototype(custodykey.KeyCreatedEvent{})
	registry.RegisterPrototype(transaction.RetriedEvent{})
}

func RegisterCrossEvents(crossRegistry *event.ListenerRegistry) {
}

func RegisterListeners(
	registry *event.ListenerRegistry,
	retryer retry.Retryer,
	bcConfig *coreconfig.BcRetryableListenerConfig,
	crossEventBus event.CrossBus,
	appServices *AppServices,
	listenerConfig *coreconfig.BcListenerConfig,
) {
	lr := shared.NewListenerRegistrar(registry, retryer, bcConfig, listenerConfig)

	lr.RegisterRetryable(keycreatedlistener.NewListener(crossEventBus))
	lr.RegisterRetryable(transitfailedtransactiontoretriedlistener.NewListener(appServices.TransitFailedTransactionToRetriedAppService))
	lr.RegisterRetryable(fundaccountlistener.NewListener(appServices.FundAccountAppService))
}
