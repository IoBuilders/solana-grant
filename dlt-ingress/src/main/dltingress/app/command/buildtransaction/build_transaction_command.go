package buildtransaction

import "dlt-ingress/src/main/dltingress/domain/transaction"

type Command struct {
	SenderDltAccountId   string
	SignersDltAccountIds []string
	SmartContractId      string
	SmartContractName    string
	MethodName           string
	MethodArgs           map[string]any
	NetworkId            string
}

type Response struct {
	transaction.TransactionBuiltEvent
}
