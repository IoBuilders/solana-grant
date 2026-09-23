package descriptor

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
)

type BroadcasterConfiguration interface {
	Descriptor[broadcaster.Configuration]
}
