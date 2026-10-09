package faucetwalletrepo

import (
	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository/base"
)

type FaucetWallet struct {
	baserepo.Model
	NetworkId        string        `gorm:"type:varchar(255) not null;uniqueIndex"`
	CustodyKeyId     uuid.UUID     `gorm:"type:uuid not null;uniqueIndex"`
	FundingAmount    amount.Amount `gorm:"type:varchar(100) not null"`
	BalanceThreshold amount.Amount `gorm:"type:varchar(100) not null"`
	Enabled          bool          `gorm:"not null;default:true"`
}
