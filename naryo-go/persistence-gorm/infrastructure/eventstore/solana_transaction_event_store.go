package eventstore

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/store"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/infrastructure"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/infrastructure/eventstore/persistence"
)

type SolanaTransactionEventStore struct {
	infrastructure.BaseStore[event.SolanaTransactionEvent]
}

func NewSolanaTransactionEventStore(db *gorm.DB) *SolanaTransactionEventStore {
	return &SolanaTransactionEventStore{
		BaseStore: infrastructure.NewBaseStore[event.SolanaTransactionEvent](db),
	}
}

func (s *SolanaTransactionEventStore) Save(ctx context.Context, data event.SolanaTransactionEvent) error {
	p, err := eventstorepersistence.FromSolanaTransactionEventDomain(data)
	if err != nil {
		return fmt.Errorf("gorm solana transaction event store: %w", err)
	}
	if err := s.DB().WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(&p).Error; err != nil {
		return fmt.Errorf("gorm solana transaction event store: %w", err)
	}

	return nil
}

var _ store.TransactionEventStore[event.SolanaTransactionEvent] = (*SolanaTransactionEventStore)(nil)
