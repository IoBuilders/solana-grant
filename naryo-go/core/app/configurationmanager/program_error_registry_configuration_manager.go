package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

type ProgramErrorRegistryConfigurationManager interface {
	ConfigurationManager[event.ProgramErrorRegistry]
}
