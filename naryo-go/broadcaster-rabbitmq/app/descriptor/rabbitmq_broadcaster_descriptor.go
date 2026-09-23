package descriptor

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/domain/broadcaster"
	coredescriptor "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
)

type RabbitMQBroadcaster interface {
	coredescriptor.Descriptor[*broadcaster.RabbitMQBroadcaster]
}
