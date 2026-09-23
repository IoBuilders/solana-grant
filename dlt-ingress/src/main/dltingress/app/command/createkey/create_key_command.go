package createkey

import "dlt-ingress/src/main/dltingress/domain/custodykey"

type Command struct {
	Dlt             string
	CustodyProvider string
}

type Response struct {
	custodykey.KeyCreatedEvent
}
