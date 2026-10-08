package txqueueslotrepo

import (
	"time"

	"github.com/google/uuid"
)

type TxQueueSlot struct {
	Id        uuid.UUID `gorm:"type:uuid;primarykey"`
	CreatedAt time.Time `gorm:"not null"`
	NetworkId string    `gorm:"type:varchar(100);not null;index:idx_slots_count,priority:1"`
	ExpiresAt time.Time `gorm:"not null;index:idx_slots_count,priority:2;index:idx_janitor_cleanup"`
}
