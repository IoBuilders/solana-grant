package descriptor

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store"
)

type Store interface {
	Descriptor[store.Configuration]
}
