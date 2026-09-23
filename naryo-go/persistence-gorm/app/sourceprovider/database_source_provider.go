package sourceprovider

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/app/descriptor"
)

type DatabaseSourceProvider = sourceprovider.SourceProvider[descriptor.Database]
