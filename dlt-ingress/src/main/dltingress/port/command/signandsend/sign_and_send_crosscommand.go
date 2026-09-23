package signandsendcross

import "dlt-ingress/src/main/dltingress/port/event/signandsend"

type CrossCommand struct {
	SenderDltAccountId   string
	SignersDltAccountIds []string
	SmartContractId      string
	SmartContractName    string
	MethodName           string
	MethodArgs           map[string]any
	NetworkId            string
}

type CrossResponse struct {
	signandsendevents.TransactionSentCrossEvent
}
