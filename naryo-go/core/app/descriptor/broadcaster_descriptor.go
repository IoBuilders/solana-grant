package descriptor

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"

type Broadcaster interface {
	Descriptor[*broadcaster.Broadcaster]
}
