package signandsendbatchcross

import "dlt-ingress/src/main/dltingress/port/crossevent/signandsend"

type CrossCommand struct {
	SenderDltAccountId   string
	SignersDltAccountIds []string
	NetworkId            string
	Calls                []BatchCall
	Dispatch             *DispatchCall // Required on EVM. Nil only on a DLT that batches natively, not supported yet.
}

type BatchCall struct {
	SmartContractId   string // Ignored on EVM: Dispatch sets the destination for the whole batch.
	SmartContractName string
	MethodName        string
	MethodArgs        map[string]any // Optional: a method may take no arguments.
}

type DispatchCall struct {
	SmartContractId   string
	SmartContractName string
	MethodName        string
	MethodArgs        map[string]any // Optional: its own arguments, the ones that are not the batch.
	CallDataArgName   string
}

type CrossResponse struct {
	signandsendevents.TransactionSentCrossEvent
}
