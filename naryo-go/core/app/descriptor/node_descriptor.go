package descriptor

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/node"

type Node interface {
	Descriptor[*node.Node]
}
