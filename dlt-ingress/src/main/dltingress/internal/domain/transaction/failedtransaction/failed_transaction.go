package failedtransaction

import (
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
)

type FailedTransaction struct {
	base.Entity
	TxId         string
	NetworkId    string
	Status       Status
	ErrorDetails string
}

func NewFailedTransaction(txId string, networkId string, errorDetails string) *FailedTransaction {
	return &FailedTransaction{
		TxId:         txId,
		NetworkId:    networkId,
		Status:       StatusNotRetried,
		ErrorDetails: errorDetails,
	}
}

type Status string

const (
	StatusNotRetried Status = "NOT_RETRIED"
	StatusRetried    Status = "RETRIED"
)

var validTypes = map[Status]struct{}{
	StatusNotRetried: {},
	StatusRetried:    {},
}

func (t Status) IsValid() bool {
	_, ok := validTypes[t]
	return ok
}

func (t Status) String() string {
	return string(t)
}

func (t Status) IsRetriedRetried() bool {
	return t == StatusRetried
}

func (t Status) IsNotRetried() bool {
	return t == StatusNotRetried
}

func (ft *FailedTransaction) TransitToRetried() error {
	if !ft.Status.IsNotRetried() {
		return domainerrors.NewEntityUnexpectedStatusDomainError("FailedTransaction", ft.Id, string(ft.Status), string(StatusNotRetried))
	}
	ft.Status = StatusRetried
	return nil
}
