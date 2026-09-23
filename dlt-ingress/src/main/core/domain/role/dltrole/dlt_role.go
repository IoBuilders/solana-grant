package dltrole

import "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/basemodel"

type DltRole struct {
	basemodel.Model
	Name        string `gorm:"type:varchar(50);unique;not null"`
	Description string `gorm:"type:varchar(255)"`
	Hash        string `gorm:"type:varchar(255);unique;not null"`
}
