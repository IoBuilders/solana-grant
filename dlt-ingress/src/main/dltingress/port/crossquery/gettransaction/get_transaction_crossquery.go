package gettransactioncross

import "dlt-ingress/src/main/dltingress/port/crossquery"

type CrossQuery struct {
	TxId string
	Dlt  string
}

type Response = *crossquery.Transaction
