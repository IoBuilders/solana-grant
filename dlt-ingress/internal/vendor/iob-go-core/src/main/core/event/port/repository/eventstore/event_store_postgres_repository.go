package eventstorerepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresEventStoreRepository struct {
	db *gorm.DB
	tm db.TransactionManager
}

func NewPostgresEventStoreRepository(db *gorm.DB, tm db.TransactionManager) *PostgresEventStoreRepository {
	return &PostgresEventStoreRepository{
		db: db,
		tm: tm,
	}
}

// Creates a new EventStore into the database.
func (r *PostgresEventStoreRepository) Create(ctx context.Context, entity *eventstore.EventStore) error {
	result := r.getDB(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(&entity)
	if result.Error != nil {
		return fmt.Errorf("failed to create EventStore: %w", result.Error)
	}
	return nil
}

// Save inserts or updates a new EventStore into the database.
func (r *PostgresEventStoreRepository) Save(ctx context.Context, entity *eventstore.EventStore) error {
	result := r.getDB(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(&entity)
	if result.Error != nil {
		return fmt.Errorf("failed to insert EventStore: %w", result.Error)
	}
	return nil
}

// Delete a EventStore from the database.
func (r *PostgresEventStoreRepository) Delete(ctx context.Context, entity *eventstore.EventStore) error {
	result := r.getDB(ctx).Delete(&entity)
	if result.Error != nil {
		return fmt.Errorf("failed to delete EventStore: %w", result.Error)
	}
	return nil
}

// FindById returns the EventStore with the given id.
func (r *PostgresEventStoreRepository) FindById(ctx context.Context, id uuid.UUID) (*eventstore.EventStore, error) {
	var entity eventstore.EventStore
	result := r.getDB(ctx).First(&entity, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewEventStoreNotFoundDomainError(id)
		}
		return nil, fmt.Errorf("failed to retrieve EventStore: %w", result.Error)
	}
	return &entity, nil
}

func (r *PostgresEventStoreRepository) ExistById(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.getDB(ctx).
		Model(&eventstore.EventStore{}).
		Select("count(*) > 0").
		Where("id = ?", id).
		Find(&exists).
		Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence of a EventStore: %w", err)
	}
	return exists, nil
}

// FindByIdPreload returns the EventStore with the given id and all associated entities.
func (r *PostgresEventStoreRepository) FindByIdPreload(ctx context.Context, id uuid.UUID) (*eventstore.EventStore, error) {
	var entity eventstore.EventStore
	result := r.getDB(ctx).Preload(clause.Associations).First(&entity, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewEventStoreNotFoundDomainError(id)
		}
		return nil, fmt.Errorf("failed to retrieve EventStore: %w", result.Error)
	}
	return &entity, nil
}

// FindAll returns all entities EventStore in the database.
func (r *PostgresEventStoreRepository) FindAll(ctx context.Context) ([]eventstore.EventStore, error) {
	var entities []eventstore.EventStore
	result := r.getDB(ctx).Find(&entities)
	if result.Error != nil {
		return entities, fmt.Errorf("failed to fetch EventStore entities: %w", result.Error)
	}
	return entities, nil
}

// FindAll returns all entities EventStore in the database including deleted.
func (r *PostgresEventStoreRepository) FindAllWithDeleted(ctx context.Context) ([]eventstore.EventStore, error) {
	var entities []eventstore.EventStore
	result := r.getDB(ctx).Unscoped().Find(&entities)
	if result.Error != nil {
		return entities, fmt.Errorf("failed to fetch EventStore entities: %w", result.Error)
	}
	return entities, nil
}

// FindAllPreload returns all entities EventStore in the database with all associated entities.
func (r *PostgresEventStoreRepository) FindAllPreload(ctx context.Context) ([]eventstore.EventStore, error) {
	var entities []eventstore.EventStore
	result := r.getDB(ctx).Preload(clause.Associations).Find(&entities)
	if result.Error != nil {
		return entities, fmt.Errorf("failed to fetch EventStore entities: %w", result.Error)
	}
	return entities, nil
}

// FindAllPaginated returns all entities EventStore in the database, paginated.
func (r *PostgresEventStoreRepository) FindAllPaginated(ctx context.Context, params pagination.PaginationParams) ([]eventstore.EventStore, error) {
	var entities []eventstore.EventStore
	result := r.getDB(ctx).Offset(params.Offset).Limit(params.PageSize).Order(params.OrderClause()).Find(&entities)
	if result.Error != nil {
		return entities, fmt.Errorf("failed to fetch EventStore entities: %w", result.Error)
	}
	return entities, nil
}

// FindAllPaginatedPreload returns all entities EventStore in the database with all associated entities, paginated.
func (r *PostgresEventStoreRepository) FindAllPaginatedPreload(ctx context.Context, params pagination.PaginationParams) ([]eventstore.EventStore, error) {
	var entities []eventstore.EventStore
	result := r.getDB(ctx).Offset(params.Offset).Limit(params.PageSize).Order(params.OrderClause()).Preload(clause.Associations).Find(&entities)
	if result.Error != nil {
		return entities, fmt.Errorf("failed to fetch EventStore entities: %w", result.Error)
	}
	return entities, nil
}

// CountAll returns the number of entities EventStore in the database.
func (r *PostgresEventStoreRepository) CountAll(ctx context.Context) (int, error) {
	var count int64
	result := r.getDB(ctx).Model(&eventstore.EventStore{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count EventStore entities: %w", result.Error)
	}
	return int(count), nil
}

func (r *PostgresEventStoreRepository) ClaimPendingEvents(ctx context.Context, isCross bool, limit int) ([]eventstore.EventStore, error) {
	var eventStores []eventstore.EventStore

	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var consumers []eventstore.EventConsumer

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ?", eventstore.Pending).
			Where("is_cross = ?", isCross).
			Order("created_at ASC").
			Limit(limit).
			Find(&consumers).Error; err != nil {
			return err
		}

		if len(consumers) == 0 {
			return nil
		}

		consumerIDs := make([]uuid.UUID, len(consumers))
		eventStoreIDs := make([]uuid.UUID, 0, len(consumers))
		seenStores := make(map[uuid.UUID]bool)

		for i, c := range consumers {
			consumerIDs[i] = c.Id
			if !seenStores[c.EventStoreId] {
				eventStoreIDs = append(eventStoreIDs, c.EventStoreId)
				seenStores[c.EventStoreId] = true
			}
		}

		if err := tx.Model(&eventstore.EventConsumer{}).
			Where("id IN ?", consumerIDs).
			Update("status", eventstore.Processing).Error; err != nil {
			return err
		}

		if err := tx.Preload("EventConsumers", "id IN ?", consumerIDs).
			Where("id IN ?", eventStoreIDs).
			Find(&eventStores).Error; err != nil {
			return err
		}

		return nil
	})

	return eventStores, err
}

func (r *PostgresEventStoreRepository) FindTraceParentByTxHash(ctx context.Context, txHash string) (string, error) {
	var traceParent string
	result := r.getDB(ctx).Model(&eventstore.EventStore{}).
		Select("trace_parent").
		Where("type LIKE ?", "%Ordered%").
		Where("payload->>'TransactionHash' = ?", txHash).
		Limit(1).
		Scan(&traceParent)
	if result.Error != nil {
		return "", fmt.Errorf("failed to find trace parent by tx hash %s: %w", txHash, result.Error)
	}
	return traceParent, nil
}

func (r *PostgresEventStoreRepository) getDB(ctx context.Context) *gorm.DB {
	val := r.tm.TransactionValue(ctx)
	if val == nil {
		return r.db.WithContext(ctx)
	}

	wrapper, ok := val.(db.GormTransaction)
	if !ok {
		return r.db.WithContext(ctx)
	}

	wrapperTx := wrapper.Tx
	txDb, errTx := wrapperTx.DB()
	repoDb, errDb := r.db.DB()

	if errTx == nil && errDb == nil && txDb == repoDb {
		return wrapperTx.WithContext(ctx)
	}

	return r.db.WithContext(ctx)
}

var _ Repository = (*PostgresEventStoreRepository)(nil)
