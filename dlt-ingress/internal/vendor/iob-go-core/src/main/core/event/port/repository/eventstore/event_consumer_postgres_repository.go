package eventstorerepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresEventConsumerRepository struct {
	db *gorm.DB
	tm db.TransactionManager
}

func NewPostgresEventConsumerRepository(db *gorm.DB, tm db.TransactionManager) *PostgresEventConsumerRepository {
	return &PostgresEventConsumerRepository{
		db: db,
		tm: tm,
	}
}

func (r *PostgresEventConsumerRepository) Create(ctx context.Context, entity *eventstore.EventConsumer) error {
	result := r.getDB(ctx).Create(&entity)
	if result.Error != nil {
		return fmt.Errorf("failed to create EventConsumer: %w", result.Error)
	}
	return nil
}

func (r *PostgresEventConsumerRepository) Save(ctx context.Context, entity *eventstore.EventConsumer) error {
	result := r.getDB(ctx).Save(&entity)
	if result.Error != nil {
		return fmt.Errorf("failed to update EventConsumer: %w", result.Error)
	}
	return nil
}

func (r *PostgresEventConsumerRepository) Delete(ctx context.Context, entity *eventstore.EventConsumer) error {
	result := r.getDB(ctx).Delete(&entity)
	if result.Error != nil {
		return fmt.Errorf("failed to delete EventConsumer: %w", result.Error)
	}
	return nil
}

func (r *PostgresEventConsumerRepository) FindById(ctx context.Context, id uuid.UUID) (*eventstore.EventConsumer, error) {
	var entity eventstore.EventConsumer
	result := r.getDB(ctx).First(&entity, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewEventConsumerNotFoundDomainError(id)
		}
		return nil, fmt.Errorf("failed to retrieve EventConsumer: %w", result.Error)
	}
	return &entity, nil
}

func (r *PostgresEventConsumerRepository) ExistById(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.getDB(ctx).
		Model(&eventstore.EventConsumer{}).
		Select("count(*) > 0").
		Where("id = ?", id).
		Find(&exists).
		Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence of a EventConsumer: %w", err)
	}
	return exists, nil
}

func (r *PostgresEventConsumerRepository) FindByIdPreload(ctx context.Context, id uuid.UUID) (*eventstore.EventConsumer, error) {
	var entity eventstore.EventConsumer
	result := r.getDB(ctx).Preload(clause.Associations).First(&entity, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewEventConsumerNotFoundDomainError(id)
		}
		return nil, fmt.Errorf("failed to retrieve EventConsumer: %w", result.Error)
	}
	return &entity, nil
}

func (r *PostgresEventConsumerRepository) FindAll(ctx context.Context) ([]eventstore.EventConsumer, error) {
	var entities []eventstore.EventConsumer
	result := r.getDB(ctx).Find(&entities)
	if result.Error != nil {
		return entities, fmt.Errorf("failed to fetch EventConsumer entities: %w", result.Error)
	}
	return entities, nil
}

func (r *PostgresEventConsumerRepository) FindAllWithDeleted(ctx context.Context) ([]eventstore.EventConsumer, error) {
	var entities []eventstore.EventConsumer
	result := r.getDB(ctx).Unscoped().Find(&entities)
	if result.Error != nil {
		return entities, fmt.Errorf("failed to fetch EventConsumer entities: %w", result.Error)
	}
	return entities, nil
}

func (r *PostgresEventConsumerRepository) FindAllPreload(ctx context.Context) ([]eventstore.EventConsumer, error) {
	var entities []eventstore.EventConsumer
	result := r.getDB(ctx).Preload(clause.Associations).Find(&entities)
	if result.Error != nil {
		return entities, fmt.Errorf("failed to fetch EventConsumer entities: %w", result.Error)
	}
	return entities, nil
}

func (r *PostgresEventConsumerRepository) FindAllPaginated(ctx context.Context, params pagination.PaginationParams) ([]eventstore.EventConsumer, error) {
	var entities []eventstore.EventConsumer
	result := r.getDB(ctx).Offset(params.Offset).Limit(params.PageSize).Order(params.OrderClause()).Find(&entities)
	if result.Error != nil {
		return entities, fmt.Errorf("failed to fetch EventConsumer entities: %w", result.Error)
	}
	return entities, nil
}

func (r *PostgresEventConsumerRepository) FindAllPaginatedPreload(ctx context.Context, params pagination.PaginationParams) ([]eventstore.EventConsumer, error) {
	var entities []eventstore.EventConsumer
	result := r.getDB(ctx).Offset(params.Offset).Limit(params.PageSize).Order(params.OrderClause()).Preload(clause.Associations).Find(&entities)
	if result.Error != nil {
		return entities, fmt.Errorf("failed to fetch EventConsumer entities: %w", result.Error)
	}
	return entities, nil
}

func (r *PostgresEventConsumerRepository) CountAll(ctx context.Context) (int, error) {
	var count int64
	result := r.getDB(ctx).Model(&eventstore.EventConsumer{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count EventConsumer entities: %w", result.Error)
	}
	return int(count), nil
}

func (r *PostgresEventConsumerRepository) UpdateStuckEventConsumers(ctx context.Context, cutoff time.Time) (int64, error) {
	var affected int64

	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var ids []uuid.UUID

		if err := tx.
			Model(&eventstore.EventConsumer{}).
			Select("id").
			Where("status = ? AND updated_at <= ?", eventstore.Processing, cutoff).
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Find(&ids).
			Error; err != nil {
			return err
		}

		if len(ids) == 0 {
			return nil
		}

		res := tx.
			Model(&eventstore.EventConsumer{}).
			Where("id IN ?", ids).
			Updates(map[string]any{
				"status":     eventstore.Pending,
				"updated_at": time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}

		affected = res.RowsAffected
		return nil
	})

	return affected, err
}

func (p *PostgresEventConsumerRepository) FindByStatusPaginated(ctx context.Context, status eventstore.Status, params pagination.PaginationParams) ([]eventstore.EventConsumer, error) {
	var entities []eventstore.EventConsumer
	result := p.getDB(ctx).Model(&eventstore.EventConsumer{}).
		Where("status = ?", status).
		Preload("EventStore").
		Limit(params.PageSize).
		Offset(params.Offset).
		Order(params.OrderClause()).
		Find(&entities)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch EventConsumer entities: %w", result.Error)
	}
	return entities, nil
}

func (p *PostgresEventConsumerRepository) CountAllByFilters(ctx context.Context, status eventstore.Status) (int, error) {
	var count int64
	result := p.getDB(ctx).Model(&eventstore.EventConsumer{}).
		Where("status = ?", status).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count EventConsumer entities: %w", result.Error)
	}
	return int(count), nil
}

func (r *PostgresEventConsumerRepository) getDB(ctx context.Context) *gorm.DB {
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

var _ EventConsumerRepository = (*PostgresEventConsumerRepository)(nil)
