package createkeycross

import (
	"dlt-ingress/src/main/dltingress/port/crossevent/custodykey"
)

type CrossCommand struct {
	Dlt string
}

type CrossResponse struct {
	custodykeyevents.KeyCreatedCrossEvent
}
