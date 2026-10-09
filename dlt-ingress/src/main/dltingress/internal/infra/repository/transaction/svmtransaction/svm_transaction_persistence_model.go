package svmtransactionrepo

import (
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/transaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type SvmTransaction struct {
	transactionrepo.Transaction
	Dlt                   common.Dlt     `gorm:"type:varchar(20) not null"`
	FeePayer              string         `gorm:"type:varchar(255) not null"`
	RecentBlockhash       string         `gorm:"type:varchar(100) not null"`
	CuLimit               *amount.Amount `gorm:"type:varchar(100) not null"`
	CuPrice               *amount.Amount `gorm:"type:varchar(100) not null"`
	SerializedTransaction string         `gorm:"type:text not null"`
}
