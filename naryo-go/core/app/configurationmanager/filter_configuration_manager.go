package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
)

type FilterConfigurationManager interface {
	CollectionConfigurationManager[filter.Filter]
}
