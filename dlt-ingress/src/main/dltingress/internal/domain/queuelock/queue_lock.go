package queuelock

import (
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
)

type QueueLock struct {
	base.Entity
	NetworkId string
}
