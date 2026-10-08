package signandsendcross

import (
	"dlt-ingress/src/main/dltingress/port/crossevent/signandsend"
)

type CrossCommand struct {
	SenderDltAccountId   string
	SignersDltAccountIds []string
	SmartContractId      string // EVM: the contract address, required. SVM: the IDL's program ID, or empty.
	SmartContractName    string
	MethodName           string
	MethodArgs           map[string]any
	NetworkId            string
}

type CrossResponse struct {
	signandsendevents.TransactionSentCrossEvent
}
