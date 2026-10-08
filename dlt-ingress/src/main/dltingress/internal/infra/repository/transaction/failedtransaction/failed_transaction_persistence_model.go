package failedtransactionrepo

import (
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/failedtransaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository/base"
)

type FailedTransaction struct {
	baserepo.Model
	TxId         string                   `gorm:"type:varchar(255) not null unique"`
	NetworkId    string                   `gorm:"type:varchar(255) not null"`
	Status       failedtransaction.Status `gorm:"type:varchar(20) not null"`
	ErrorDetails string                   `gorm:"type:varchar(5000) not null"`
}
