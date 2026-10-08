package queuelockrepo

import (
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository/base"
)

type QueueLock struct {
	baserepo.Model
	NetworkId string `gorm:"type:varchar(100);not null;index:idx_network_queue,unique"`
}
