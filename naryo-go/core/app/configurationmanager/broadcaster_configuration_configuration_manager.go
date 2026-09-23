package configurationmanager

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
)

type BroadcasterConfigurationConfigurationManager interface {
	CollectionConfigurationManager[broadcaster.Configuration]
	RegisterMapper(ctx context.Context, configurationType string, mapper func(ctx context.Context, source broadcaster.Configuration) (broadcaster.Configuration, error)) error
}
