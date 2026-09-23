package repository

import (
	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/core/utils"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gorm.io/gorm"
)

func NewPostgresDB() (*gorm.DB, error) {
	return utils.NewPostgresDB(config.AppConfig.Database.DltIngress)
}

// Migrate applies all SQL migrations from the predefined directory
func Migrate(gormdb *gorm.DB) error {
	return db.Migrate(gormdb, "dltingress")
}
