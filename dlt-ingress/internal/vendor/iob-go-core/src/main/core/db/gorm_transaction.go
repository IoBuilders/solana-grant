package db

import (
	"gorm.io/gorm"
)

type GormTransaction struct {
	Tx *gorm.DB
}

func NewGormTransaction(db *gorm.DB) GormTransaction {
	return GormTransaction{
		Tx: db.Begin(),
	}
}

func (t GormTransaction) Commit() error {
	return t.Tx.Commit().Error

}

func (t GormTransaction) Rollback() error {
	return t.Tx.Rollback().Error
}
