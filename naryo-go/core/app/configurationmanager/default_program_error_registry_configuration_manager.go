package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

type DefaultProgramErrorRegistryConfigurationManager struct {
	BaseConfigurationManager[event.ProgramErrorRegistry, descriptor.ProgramErrorRegistry]
}

func NewDefaultProgramErrorRegistryConfigurationManager(providers []sourceprovider.ProgramErrorRegistrySourceProvider) *DefaultProgramErrorRegistryConfigurationManager {
	return &DefaultProgramErrorRegistryConfigurationManager{
		BaseConfigurationManager: NewBaseConfigurationManager(providers),
	}
}
