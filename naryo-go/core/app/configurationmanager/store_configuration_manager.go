package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store"
)

type StoreConfigurationManager interface {
	CollectionConfigurationManager[store.Configuration]
}
