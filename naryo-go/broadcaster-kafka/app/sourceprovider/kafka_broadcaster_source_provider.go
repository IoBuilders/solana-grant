package sourceprovider

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/app/descriptor"
	coresourceprovider "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
)

type KafkaBroadcasterSourceProvider = coresourceprovider.SourceProvider[descriptor.KafkaBroadcaster]
