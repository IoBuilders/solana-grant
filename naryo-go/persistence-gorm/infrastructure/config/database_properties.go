package config

import (
	"time"

	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/domain/database"
)

type DatabaseProperties struct {
	Url                             string        `mapstructure:"url"`
	MaxOpenConnections              int           `mapstructure:"maxOpenConnections"`
	MaxIdleConnections              int           `mapstructure:"maxIdleConnections"`
	ConnMaxLifetime                 time.Duration `mapstructure:"connMaxLifetime"`
	ConnMaxIdleTime                 time.Duration `mapstructure:"connMaxIdleTime"`
	ConnectTimeout                  time.Duration `mapstructure:"connectTimeout"`
	StatementTimeout                time.Duration `mapstructure:"statementTimeout"`
	IdleInTransactionSessionTimeout time.Duration `mapstructure:"idleInTransactionSessionTimeout"`
	LockTimeout                     time.Duration `mapstructure:"lockTimeout"`
}

func (d *DatabaseProperties) Map() (*database.Database, error) {
	return database.NewDatabase(d.Url, d.MaxOpenConnections, d.MaxIdleConnections, d.ConnMaxLifetime,
		d.ConnMaxIdleTime, d.ConnectTimeout, d.StatementTimeout, d.IdleInTransactionSessionTimeout, d.LockTimeout)
}

var _ descriptor.Database = (*DatabaseProperties)(nil)
