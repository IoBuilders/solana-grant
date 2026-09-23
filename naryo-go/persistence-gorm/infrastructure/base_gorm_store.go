package infrastructure

import (
	"reflect"

	"gorm.io/gorm"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store"
)

type BaseStore[D any] struct {
	db *gorm.DB
}

func NewBaseStore[D any](db *gorm.DB) BaseStore[D] {
	return BaseStore[D]{db: db}
}

func (s BaseStore[D]) DB() *gorm.DB {
	return s.db
}

func (s BaseStore[D]) Supports(storeType store.Type, dataType reflect.Type) bool {
	return storeType == store.TypeGorm && dataType == reflect.TypeFor[D]()
}
