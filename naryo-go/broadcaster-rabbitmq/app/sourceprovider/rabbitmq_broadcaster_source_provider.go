package sourceprovider

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/app/descriptor"
	coresourceprovider "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
)

type RabbitMQBroadcasterSourceProvider = coresourceprovider.SourceProvider[descriptor.RabbitMQBroadcaster]
