package base

import "github.com/google/uuid"

type UnsignedTransactionProcess struct {
	Id           uuid.UUID
	ProcessId    uuid.UUID
	DltAccountId string
	Payload      string
}
