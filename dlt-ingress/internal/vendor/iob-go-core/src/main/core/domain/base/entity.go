package base

import (
	"time"

	"github.com/google/uuid"
)

type Entity struct {
	Id        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewEntity() Entity {
	return Entity{
		Id:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
