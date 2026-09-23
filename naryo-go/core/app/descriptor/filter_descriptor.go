package descriptor

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"

type Filter interface {
	Descriptor[filter.Filter]
}
