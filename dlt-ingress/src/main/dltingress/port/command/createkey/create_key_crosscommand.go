package createkeycross

import "dlt-ingress/src/main/dltingress/port/event/custodykey"

type CrossCommand struct {
	Dlt string
}

type CrossResponse struct {
	custodykeyevents.KeyCreatedCrossEvent
}
