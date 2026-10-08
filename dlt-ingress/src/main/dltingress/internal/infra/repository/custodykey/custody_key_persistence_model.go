package custodykeyrepo

import (
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository/base"
)

type CustodyKey struct {
	baserepo.Model
	KeyType         common.KeyType             `gorm:"type:varchar(30) not null"`
	Status          custodykey.Status          `gorm:"type:varchar(20) not null;default:ACTIVE"`
	DltAccountId    string                     `gorm:"type:varchar(255) not null;uniqueIndex"`
	Dlt             common.Dlt                 `gorm:"type:varchar(20) not null"`
	ExternalId      string                     `gorm:"type:varchar(255) not null"`
	CustodyProvider custodykey.CustodyProvider `gorm:"type:varchar(30) not null"`
}
