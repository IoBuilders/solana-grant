package healthchecks

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/health"
	"gorm.io/gorm"
)

type DatabaseChecker struct {
	name string
	db   *gorm.DB
}

func NewDatabaseChecker(name string, db *gorm.DB) *DatabaseChecker {
	return &DatabaseChecker{
		name: name,
		db:   db,
	}
}

func (d *DatabaseChecker) Name() string {
	return d.name
}

func (d *DatabaseChecker) Check(ctx context.Context) health.CheckResult {
	start := time.Now()
	sqlDB, err := d.db.DB()
	if err != nil {
		return health.NewDownCheckResult(fmt.Errorf("failed to get sql.DB instance: %w", err), start)
	}
	err = sqlDB.PingContext(ctx)
	if err != nil {
		return health.NewDownCheckResult(fmt.Errorf("ping failed: %w", err), start)
	}
	return health.NewUpCheckResult(start)
}
