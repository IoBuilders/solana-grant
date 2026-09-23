package modelprocess

import "github.com/google/uuid"

type UnsignedTransactionProcessModel struct {
	Id           uuid.UUID `json:"id"`
	ProcessId    uuid.UUID `json:"processId"`
	DltAccountId string    `json:"dltAccountId"`
	Payload      string    `json:"payload"`
}

type BaseProcessModel struct {
	Id                         uuid.UUID                        `json:"id"`
	Status                     string                           `json:"status"`
	Type                       string                           `json:"type"`
	EntityId                   *uuid.UUID                       `json:"entityId,omitempty"`
	SigningType                string                           `json:"signingType,omitempty"`
	SignerDltAccountId         string                           `json:"signerDltAccountId,omitempty"`
	TransactionHash            string                           `json:"transactionHash,omitempty"`
	UnsignedTransactionProcess *UnsignedTransactionProcessModel `json:"unsignedTransactionProcess,omitempty"`
}

type DerivedProcessModel struct {
	Id            uuid.UUID `json:"id"`
	Status        string    `json:"status"`
	Type          string    `json:"type"`
	EntityId      uuid.UUID `json:"entityId"`
	BaseProcessId uuid.UUID `json:"baseProcessId"`
}
