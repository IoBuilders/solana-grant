package event

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Bus interface {
	Publish(ctx context.Context, event Event) error
}

type Event interface {
	GetId() uuid.UUID
	GetCreatedAt() time.Time
}

func NewBaseEvent() *BaseEvent {
	return &BaseEvent{
		Id:        uuid.New(),
		CreatedAt: time.Now(),
	}
}

type BaseEvent struct {
	Id        uuid.UUID
	CreatedAt time.Time
}

func (e BaseEvent) GetId() uuid.UUID {
	return e.Id
}

func (e BaseEvent) GetCreatedAt() time.Time {
	return e.CreatedAt
}

type DltEvent interface {
	Event
	GetTransactionHash() string
	GetTimeStamp() time.Time
	GetBlockTimestamp() time.Time
}

type BaseDltEvent struct {
	BaseEvent
	TransactionHash         string
	OriginalTransactionHash *string
	TransactionIndex        uint64
	BlockNumber             uint64
	BlockHash               string
	BlockTimestamp          time.Time
	DltAccountId            string
	Signer                  string
	TransactionOrigin       string
	NetworkId               string
}

func NewBaseDltEvent() *BaseDltEvent {
	return &BaseDltEvent{
		BaseEvent: BaseEvent{
			Id:        uuid.New(),
			CreatedAt: time.Now(),
		},
	}
}

func (b *BaseDltEvent) GetTransactionHash() string { return b.TransactionHash }
func (b *BaseDltEvent) GetBlockNumber() uint64     { return b.BlockNumber }
func (b *BaseDltEvent) GetBlockHash() string       { return b.BlockHash }
