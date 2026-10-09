package gettransaction

import "dlt-ingress/src/main/dltingress/internal/app/query"

type Query struct {
	TxId string
	Dlt  string
}

type Response = query.Transaction
