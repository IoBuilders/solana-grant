package config

import (
	"fmt"
	"time"
)

func StartCallbackServiceWorker(cll func(interval time.Duration)) {
	interval, err := time.ParseDuration(AppConfig.Server.CallbackInterval)

	if err != nil {
		panic(fmt.Sprintf("failure to parse duration for callback interval: %s", err))
	}

	cll(interval)
}
