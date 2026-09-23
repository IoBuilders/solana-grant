package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/domain/database"
)

type HttpClientConfigurationManager interface {
	configurationmanager.ConfigurationManager[*database.Database]
}
