package getcustodykey

import "dlt-ingress/src/main/dltingress/internal/app/query"

type Query struct {
	DltAccountId string
}

type Response = query.CustodyKey
