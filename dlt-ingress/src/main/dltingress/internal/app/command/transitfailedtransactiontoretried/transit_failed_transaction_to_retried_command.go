package transitfailedtransactiontoretried

import (
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/failedtransaction"
)

type Command struct {
	TxId string
}

type Response struct {
	failedtransaction.TransitFailedTransactionToRetriedEvent
}
