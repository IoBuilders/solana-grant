package descriptor

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/domain/database"
)

type Database interface {
	descriptor.Descriptor[*database.Database]
}
