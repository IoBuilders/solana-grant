package signandsend

import "dlt-ingress/src/main/dltingress/internal/domain/transaction"

type Command struct {
	SenderDltAccountId   string
	SignersDltAccountIds []string
	SmartContractId      string
	SmartContractName    string
	MethodName           string
	MethodArgs           map[string]any
	NetworkId            string
	ResolveNestedCalls   bool
}

type Response struct {
	transaction.TransactionSentEvent
}
