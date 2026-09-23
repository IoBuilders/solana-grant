package descriptor

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"

type ProgramErrorRegistry interface {
	Descriptor[event.ProgramErrorRegistry]
}
