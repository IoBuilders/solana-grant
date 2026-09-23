package eventstore

import (
	"context"
	"fmt"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/store"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/infrastructure"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/infrastructure/eventstore/persistence"
	"gorm.io/gorm"
)

type SolanaContractEventStore struct {
	infrastructure.BaseStore[event.SolanaContractEvent]
}

func NewSolanaContractEventStore(db *gorm.DB) *SolanaContractEventStore {
	return &SolanaContractEventStore{
		BaseStore: infrastructure.NewBaseStore[event.SolanaContractEvent](db),
	}
}

func (s *SolanaContractEventStore) Save(ctx context.Context, data event.SolanaContractEvent) error {
	p, err := eventstorepersistence.FromSolanaContractEventDomain(data)
	if err != nil {
		return fmt.Errorf("gorm solana contract event store: %w", err)
	}
	if err := s.DB().WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(&p).Error; err != nil {
		return fmt.Errorf("gorm solana contract event store: %w", err)
	}

	return nil
}

var _ store.ContractEventStore[event.SolanaContractEvent] = (*SolanaContractEventStore)(nil)
