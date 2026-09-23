package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/domain/database"
)

type DefaultDatabaseConfigurationManager struct {
	configurationmanager.BaseConfigurationManager[*database.Database, descriptor.Database]
}

func NewDefaultDatabaseConfigurationManager(providers []sourceprovider.DatabaseSourceProvider) *DefaultDatabaseConfigurationManager {
	return &DefaultDatabaseConfigurationManager{
		BaseConfigurationManager: configurationmanager.NewBaseConfigurationManager(providers),
	}
}
