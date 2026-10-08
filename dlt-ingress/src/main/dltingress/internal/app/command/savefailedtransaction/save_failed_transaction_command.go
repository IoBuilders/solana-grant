package savefailedtransaction

import "dlt-ingress/src/main/dltingress/internal/domain/transaction/failedtransaction"

type Command struct {
	TxId         string
	NetworkId    string
	ErrorDetails string
}

type Response struct {
	failedtransaction.SavedEvent
}
