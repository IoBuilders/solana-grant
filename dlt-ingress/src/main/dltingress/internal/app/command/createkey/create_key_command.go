package createkey

import "dlt-ingress/src/main/dltingress/internal/domain/custodykey"

type Command struct {
	Dlt             string
	CustodyProvider string
}

type Response struct {
	custodykey.KeyCreatedEvent
}
