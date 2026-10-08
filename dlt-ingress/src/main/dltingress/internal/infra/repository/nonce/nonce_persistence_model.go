package noncerepo

import (
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository/base"
)

type Nonce struct {
	baserepo.Model
	DltAccountId string         `gorm:"type:varchar(255) not null;uniqueIndex:idx_nonce_dltaccount_network"`
	NetworkId    string         `gorm:"type:varchar(255) not null;uniqueIndex:idx_nonce_dltaccount_network"`
	Value        *amount.Amount `gorm:"type:varchar(100) not null"`
}
