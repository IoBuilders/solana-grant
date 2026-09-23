package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/httpclient"
)

type HttpClientConfigurationManager interface {
	ConfigurationManager[*httpclient.HttpClient]
}
