package eventstore

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	appstore "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/store"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/infrastructure"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/infrastructure/eventstore/persistence"
)

type SolanaBlockEventStore struct {
	infrastructure.BaseStore[event.SolanaBlockEvent]
}

func NewSolanaBlockEventStore(db *gorm.DB) *SolanaBlockEventStore {
	return &SolanaBlockEventStore{
		BaseStore: infrastructure.NewBaseStore[event.SolanaBlockEvent](db),
	}
}

func (s *SolanaBlockEventStore) Save(ctx context.Context, data event.SolanaBlockEvent) error {
	p, err := eventstorepersistence.FromSolanaBlockEventDomain(data)
	if err != nil {
		return fmt.Errorf("gorm solana block event store: %w", err)
	}
	if err := s.DB().WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(&p).Error; err != nil {
		return fmt.Errorf("gorm solana block event store: %w", err)
	}

	return nil
}

func (s *SolanaBlockEventStore) GetLatest(ctx context.Context, nodeID uuid.UUID) (uint64, error) {
	var latestBlock eventstorepersistence.SolanaBlockEvent
	if err := s.DB().WithContext(ctx).
		Where("node_id = ?", nodeID).
		Order("slot DESC").First(&latestBlock).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return latestBlock.Slot, nil
}

var _ appstore.BlockEventStore[event.SolanaBlockEvent] = (*SolanaBlockEventStore)(nil)
