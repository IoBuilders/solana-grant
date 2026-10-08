package evmtransactionrepo

import (
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/transaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type EvmTransaction struct {
	transactionrepo.Transaction
	Dlt                  common.Dlt     `gorm:"type:varchar(20) not null"`
	FromAddress          string         `gorm:"type:varchar(255) not null"`
	ToAddress            string         `gorm:"type:varchar(255) not null"`
	Nonce                *amount.Amount `gorm:"type:varchar(100) not null"`
	Value                *amount.Amount `gorm:"type:varchar(100) not null"`
	TransactionType      uint           `gorm:"not null"`
	GasLimit             *amount.Amount `gorm:"type:varchar(100) not null"`
	GasPrice             *amount.Amount `gorm:"type:varchar(100)"`
	MaxPriorityFeePerGas *amount.Amount `gorm:"type:varchar(100)"`
	MaxFeePerGas         *amount.Amount `gorm:"type:varchar(100)"`
	Data                 string         `gorm:"type:text"`
}
