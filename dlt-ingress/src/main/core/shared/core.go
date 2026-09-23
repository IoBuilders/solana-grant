package shared

import (
	"dlt-ingress/src/main/core/blockchain/client"
	"dlt-ingress/src/main/dltingress/port/evm"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/cache"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/health"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/retry"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
	"go.opentelemetry.io/otel/trace"
)

type CoreCommon struct {
	QueryBus              query.Bus
	CrossQueryBus         query.CrossBus
	CrossCommandBus       command.CrossBus
	CrossEventBus         event.CrossBus
	CrossListenerRegistry *event.ListenerRegistry
	BlockchainClient      clientblockchain.Service
	EvmClientRegistry     evm.ClientRegistry
	Retryer               retry.Retryer
	Tracer                trace.Tracer
	HealthRegistry        *health.Registry
	Cache                 cache.Port
	MetricsRegistry       *metrics.Registry
}
