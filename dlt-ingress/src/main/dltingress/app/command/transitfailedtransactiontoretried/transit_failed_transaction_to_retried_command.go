package transitfailedtransactiontoretried

import (
	"dlt-ingress/src/main/dltingress/domain/transaction/failedtransaction"
)

type Command struct {
	TxId string
}

type Response struct {
	failedtransaction.TransitFailedTransactionToRetriedEvent
}
