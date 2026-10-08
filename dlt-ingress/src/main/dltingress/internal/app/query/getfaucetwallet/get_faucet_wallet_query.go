package getfaucetwallet

import (
	"dlt-ingress/src/main/dltingress/internal/app/query"
)

type Query struct {
	NetworkId string
}

type Response = query.FaucetWallet
