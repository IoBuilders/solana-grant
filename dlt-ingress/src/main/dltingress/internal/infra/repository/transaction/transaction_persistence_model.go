package transactionrepo

import (
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository/base"
)

type Transaction struct {
	baserepo.Model
	TxId         string  `gorm:"type:varchar(255) not null unique"`
	OriginalTxId *string `gorm:"type:varchar(255)"`
	NetworkId    string  `gorm:"type:varchar(255) not null"`
	NetworkUrl   string  `gorm:"type:varchar(255) not null"`
}
