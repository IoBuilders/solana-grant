package baseprocess

import (
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/basemodel"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/callback"

	"github.com/google/uuid"
)

type ProcessStatus string

const (
	Started                       ProcessStatus = "STARTED"
	Ordered                       ProcessStatus = "ORDERED"
	Created                       ProcessStatus = "CREATED"
	Finished                      ProcessStatus = "FINISHED"
	Failed                        ProcessStatus = "FAILED"
	TransactionBuilt              ProcessStatus = "TRANSACTION_BUILT"
	AwaitingSignedTransaction     ProcessStatus = "AWAITING_SIGNED_TRANSACTION"
	AllocationInstructionCreated  ProcessStatus = "ALLOCATION_INSTRUCTION_CREATED"
	AllocationInstructionRejected ProcessStatus = "ALLOCATION_INSTRUCTION_REJECTED"
)

var validStatus = map[ProcessStatus]struct{}{
	Started:                       {},
	Ordered:                       {},
	Created:                       {},
	Finished:                      {},
	Failed:                        {},
	TransactionBuilt:              {},
	AwaitingSignedTransaction:     {},
	AllocationInstructionCreated:  {},
	AllocationInstructionRejected: {},
}

func (t ProcessStatus) IsValid() bool {
	_, ok := validStatus[t]
	return ok
}

type Type string

type SigningType string

const (
	Custodial    SigningType = "CUSTODIAL"
	NonCustodial SigningType = "NON_CUSTODIAL"
)

// Deprecated: Migrate to split domain & persistence entities
type UnsignedTransactionProcess struct {
	Id           uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProcessId    uuid.UUID `gorm:"type:uuid;not null"`
	DltAccountId string    `gorm:"type:varchar(255);not null"`
	Payload      string    `gorm:"type:text;not null"`
}

// Deprecated: Migrate to split domain & persistence entities
type BaseProcess struct {
	basemodel.Model
	Status                     ProcessStatus               `gorm:"type:varchar(20)"`
	Type                       Type                        `gorm:"type:varchar(100);not null"`
	SigningType                SigningType                 `gorm:"type:varchar(20);default:CUSTODIAL"`
	EntityId                   uuid.UUID                   `gorm:"type:uuid"`
	TransactionHash            string                      `gorm:"type:varchar(255)"`
	SignerDltAccountId         string                      `gorm:"type:varchar(255);not null"`
	Callback                   callback.Callback           `gorm:"foreignKey:Id"`
	UnsignedTransactionProcess *UnsignedTransactionProcess `gorm:"foreignKey:ProcessId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}
