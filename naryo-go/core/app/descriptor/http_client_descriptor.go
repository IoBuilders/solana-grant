package descriptor

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/httpclient"

type HttpClient interface {
	Descriptor[*httpclient.HttpClient]
}
