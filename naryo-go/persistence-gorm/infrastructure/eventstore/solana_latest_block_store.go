package eventstore

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	appstore "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/store"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/infrastructure"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/infrastructure/eventstore/persistence"
)

type SolanaLatestBlockStore struct {
	infrastructure.BaseStore[event.SolanaBlockEvent]
}

func NewSolanaLatestBlockStore(db *gorm.DB) *SolanaLatestBlockStore {
	return &SolanaLatestBlockStore{
		BaseStore: infrastructure.NewBaseStore[event.SolanaBlockEvent](db),
	}
}

func (s *SolanaLatestBlockStore) Save(ctx context.Context, data event.SolanaBlockEvent) error {
	p, err := eventstorepersistence.FromSolanaLatestBlockDomain(data)
	if err != nil {
		return fmt.Errorf("gorm solana latest block store: %w", err)
	}
	if err := s.DB().WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "node_id"}}, UpdateAll: true}).Create(&p).Error; err != nil {
		return fmt.Errorf("gorm solana latest block store: %w", err)
	}

	return nil
}

func (s *SolanaLatestBlockStore) Get(ctx context.Context, nodeID uuid.UUID) (uint64, error) {
	var latestBlock eventstorepersistence.SolanaLatestBlock
	if err := s.DB().WithContext(ctx).Where("node_id = ?", nodeID).Take(&latestBlock).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, fmt.Errorf("gorm solana latest block store: %w", err)
	}
	return latestBlock.Slot, nil
}

var _ appstore.LatestBlockStore[event.SolanaBlockEvent] = (*SolanaLatestBlockStore)(nil)
