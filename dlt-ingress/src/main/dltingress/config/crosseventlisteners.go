package dltingressconfig

import (
	"dlt-ingress/src/main/config"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/retry"
)

func RegisterCrossListeners(
	crossRegistry *event.ListenerRegistry,
	retryer retry.Retryer,
	bcConfig *config.BcRetryableListenerConfig,
	listenerConfig *config.BcListenerConfig,
) {
}
