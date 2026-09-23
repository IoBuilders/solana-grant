package database

import (
	"time"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

type Database struct {
	Url                             string
	MaxOpenConnections              int
	MaxIdleConnections              int
	ConnMaxLifetime                 time.Duration
	ConnMaxIdleTime                 time.Duration
	ConnectTimeout                  time.Duration
	StatementTimeout                time.Duration
	IdleInTransactionSessionTimeout time.Duration
	LockTimeout                     time.Duration
}

func defaultDatabase() *Database {
	return &Database{
		MaxOpenConnections:              100,
		MaxIdleConnections:              10,
		ConnMaxLifetime:                 time.Hour,
		ConnMaxIdleTime:                 30 * time.Minute,
		ConnectTimeout:                  10 * time.Second,
		StatementTimeout:                60 * time.Second,
		IdleInTransactionSessionTimeout: 60 * time.Second,
		LockTimeout:                     5 * time.Second,
	}
}

func NewDatabase(
	url string,
	maxOpenConnections int,
	maxIdleConnections int,
	connMaxLifetime time.Duration,
	connMaxIdleTime time.Duration,
	connectTimeout time.Duration,
	statementTimeout time.Duration,
	idleInTransactionSessionTimeout time.Duration,
	lockTimeout time.Duration,
) (*Database, error) {
	if url == "" {
		return nil, domainerrors.NewEmptyFieldError("Url", "Database")
	}
	db := defaultDatabase()
	db.Url = url
	if maxOpenConnections > 0 {
		db.MaxOpenConnections = maxOpenConnections
	}
	if maxIdleConnections > 0 {
		db.MaxIdleConnections = maxIdleConnections
	}
	if connMaxLifetime > 0 {
		db.ConnMaxLifetime = connMaxLifetime
	}
	if connMaxIdleTime > 0 {
		db.ConnMaxIdleTime = connMaxIdleTime
	}
	if connectTimeout > 0 {
		db.ConnectTimeout = connectTimeout
	}
	if statementTimeout > 0 {
		db.StatementTimeout = statementTimeout
	}
	if idleInTransactionSessionTimeout > 0 {
		db.IdleInTransactionSessionTimeout = idleInTransactionSessionTimeout
	}
	if lockTimeout > 0 {
		db.LockTimeout = lockTimeout
	}
	return db, nil
}
