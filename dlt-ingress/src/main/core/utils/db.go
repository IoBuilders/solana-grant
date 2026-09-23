package utils

import (
	"dlt-ingress/src/main/config"
	"errors"
	"strings"

	"github.com/lib/pq"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gorm.io/gorm"
)

func NewPostgresDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var options []db.DBOption

	dbURL := ensureTimezoneUTC(cfg.Url)

	if cfg.MaxOpenConnections > 0 {
		options = append(options, db.WithMaxOpenConnections(cfg.MaxOpenConnections))
	}
	if cfg.MaxIdleConnections > 0 {
		options = append(options, db.WithMaxIdleConnections(cfg.MaxIdleConnections))
	}
	if cfg.ConnMaxLifetime > 0 {
		options = append(options, db.WithConnMaxLifetime(cfg.ConnMaxLifetime))
	}
	if cfg.ConnMaxIdleTime > 0 {
		options = append(options, db.WithConnMaxIdleTime(cfg.ConnMaxIdleTime))
	}
	if cfg.ConnectTimeout > 0 {
		options = append(options, db.WithConnectTimeout(cfg.ConnectTimeout))
	}
	if cfg.StatementTimeout > 0 {
		options = append(options, db.WithStatementTimeout(cfg.StatementTimeout))
	}
	if cfg.IdleInTransactionSessionTimeout > 0 {
		options = append(options, db.WithIdleInTransactionSessionTimeout(cfg.IdleInTransactionSessionTimeout))
	}
	if cfg.LockTimeout > 0 {
		options = append(options, db.WithLockTimeout(cfg.LockTimeout))
	}

	return db.NewPostgresDB(dbURL, options...)
}

const pgErrCodeLockTimeout = "55P03"

func IsLockTimeoutError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && string(pqErr.Code) == pgErrCodeLockTimeout
}

func ensureTimezoneUTC(dbURL string) string {
	questionMarkIndex := strings.Index(dbURL, "?")

	if questionMarkIndex == -1 {
		return dbURL + "?timezone=UTC"
	}

	hashIndex := strings.Index(dbURL[questionMarkIndex:], "#")
	queryEndIndex := len(dbURL)
	if hashIndex != -1 {
		queryEndIndex = questionMarkIndex + hashIndex
	}

	rawQuery := dbURL[questionMarkIndex+1 : queryEndIndex]
	parts := strings.Split(rawQuery, "&")
	normalizedParams := make([]string, 0, len(parts)+1)
	timezoneFound := false

	for _, param := range parts {
		if param == "" {
			continue
		}

		key := param
		if eqIndex := strings.Index(param, "="); eqIndex != -1 {
			key = param[:eqIndex]
		}

		if strings.EqualFold(key, "timezone") {
			if !timezoneFound {
				normalizedParams = append(normalizedParams, "timezone=UTC")
				timezoneFound = true
			}
			continue
		}

		normalizedParams = append(normalizedParams, param)
	}

	if !timezoneFound {
		normalizedParams = append(normalizedParams, "timezone=UTC")
	}

	return dbURL[:questionMarkIndex+1] + strings.Join(normalizedParams, "&") + dbURL[queryEndIndex:]
}
