package baserepo

import (
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/callback"

	"github.com/google/uuid"
)

type Process struct {
	Model
	ParentId                   *uuid.UUID                  `gorm:"type:uuid;index"`
	Status                     base.ProcessStatus          `gorm:"type:varchar(20)"`
	Type                       base.Type                   `gorm:"type:varchar(100);not null"`
	SigningType                base.SigningType            `gorm:"type:varchar(20);default:CUSTODIAL"`
	EntityId                   uuid.UUID                   `gorm:"type:uuid"`
	TransactionHash            string                      `gorm:"type:varchar(255)"`
	SignerDltAccountId         string                      `gorm:"type:varchar(255);not null"`
	RevertReason               string                      `gorm:"type:varchar(5000)"`
	Callback                   callback.Callback           `gorm:"foreignKey:Id"`
	UnsignedTransactionProcess *UnsignedTransactionProcess `gorm:"foreignKey:ProcessId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Subprocesses               []Process                   `gorm:"foreignKey:ParentId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;<-:false"`
}

func (Process) TableName() string {
	return "base_processes"
}
