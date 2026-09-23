package roleassociation

import (
	"dlt-ingress/src/main/core/domain/role/dltrole"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/basemodel"
)

type RoleAssociation struct {
	basemodel.Model
	AccountId uuid.UUID       `gorm:"type:uuid;not null"`
	DltRoleId uuid.UUID       `gorm:"type:uuid;not null;index"`
	DltRole   dltrole.DltRole `gorm:"foreignKey:DltRoleId;references:Id"`
}
