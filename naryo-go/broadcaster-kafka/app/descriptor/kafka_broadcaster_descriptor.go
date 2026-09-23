package descriptor

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/domain/broadcaster"
)
import coredescriptor "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"

type KafkaBroadcaster interface {
	coredescriptor.Descriptor[*broadcaster.KafkaBroadcaster]
}
