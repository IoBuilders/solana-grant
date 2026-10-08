package model

import (
	"time"

	"github.com/google/uuid"
)

type ContractEvent struct {
	TargetUrls []string
}

func NewContractEvent(targetUrls []string) ContractEvent {
	return ContractEvent{
		TargetUrls: targetUrls,
	}
}

func (ce ContractEvent) GetId() uuid.UUID {
	return uuid.New()
}

func (ce ContractEvent) GetCreatedAt() time.Time {
	return time.Now()
}
