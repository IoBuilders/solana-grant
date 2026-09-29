package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBConfig holds configuration for the database connection pool and timeouts
type DBConfig struct {
	MaxOpenConnections              int
	MaxIdleConnections              int
	ConnMaxLifetime                 time.Duration
	ConnMaxIdleTime                 time.Duration
	ConnectTimeout                  time.Duration
	StatementTimeout                time.Duration
	IdleInTransactionSessionTimeout time.Duration
	LockTimeout                     time.Duration
}

type DBOption func(config *DBConfig)

func WithMaxOpenConnections(maxOpenConnections int) DBOption {
	return func(config *DBConfig) {
		config.MaxOpenConnections = maxOpenConnections
	}
}

func WithMaxIdleConnections(maxIdleConnections int) DBOption {
	return func(config *DBConfig) {
		config.MaxIdleConnections = maxIdleConnections
	}
}

func WithConnMaxLifetime(connMaxLifetime time.Duration) DBOption {
	return func(config *DBConfig) {
		config.ConnMaxLifetime = connMaxLifetime
	}
}

func WithConnMaxIdleTime(connMaxIdleTime time.Duration) DBOption {
	return func(config *DBConfig) {
		config.ConnMaxIdleTime = connMaxIdleTime
	}
}

func WithConnectTimeout(connectTimeout time.Duration) DBOption {
	return func(config *DBConfig) {
		config.ConnectTimeout = connectTimeout
	}
}

func WithStatementTimeout(statementTimeout time.Duration) DBOption {
	return func(config *DBConfig) {
		config.StatementTimeout = statementTimeout
	}
}

func WithIdleInTransactionSessionTimeout(idleInTransactionSessionTimeout time.Duration) DBOption {
	return func(config *DBConfig) {
		config.IdleInTransactionSessionTimeout = idleInTransactionSessionTimeout
	}
}

func WithLockTimeout(lockTimeout time.Duration) DBOption {
	return func(config *DBConfig) {
		config.LockTimeout = lockTimeout
	}
}

// DefaultDBConfig returns a DBConfig with industry-standard production values
func DefaultDBConfig() DBConfig {
	return DBConfig{
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

// NewPostgresDB creates a new DB connection with the given DSN and optional configuration
func NewPostgresDB(dbURL string, opts ...DBOption) (*gorm.DB, error) {
	if dbURL == "" {
		return nil, fmt.Errorf("database Url not set")
	}

	config := DefaultDBConfig()
	for _, opt := range opts {
		opt(&config)
	}

	logger.Info("📦 Connecting to database")

	// Add timeouts to DSN if they are not already present
	var params []string
	if !strings.Contains(dbURL, "connect_timeout=") && config.ConnectTimeout > 0 {
		// connect_timeout is in seconds
		params = append(params, fmt.Sprintf("connect_timeout=%d", int(config.ConnectTimeout.Seconds())))
	}
	if !strings.Contains(dbURL, "statement_timeout=") && config.StatementTimeout > 0 {
		// statement_timeout is in milliseconds
		params = append(params, fmt.Sprintf("statement_timeout=%d", config.StatementTimeout.Milliseconds()))
	}
	if !strings.Contains(dbURL, "idle_in_transaction_session_timeout=") && config.IdleInTransactionSessionTimeout > 0 {
		// idle_in_transaction_session_timeout is in milliseconds
		params = append(params, fmt.Sprintf("idle_in_transaction_session_timeout=%d", config.IdleInTransactionSessionTimeout.Milliseconds()))
	}
	if !strings.Contains(dbURL, "lock_timeout=") && config.LockTimeout > 0 {
		// lock_timeout is in milliseconds
		params = append(params, fmt.Sprintf("lock_timeout=%d", config.LockTimeout.Milliseconds()))
	}

	if len(params) > 0 {
		separator := "?"
		if strings.Contains(dbURL, "?") {
			separator = "&"
		}
		dbURL = fmt.Sprintf("%s%s%s", dbURL, separator, strings.Join(params, "&"))
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dbURL,
	}), &gorm.Config{
		Logger: newDatabaseLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Set connection pool settings
	if config.MaxOpenConnections > 0 {
		sqlDB.SetMaxOpenConns(config.MaxOpenConnections)
	}
	if config.MaxIdleConnections > 0 {
		sqlDB.SetMaxIdleConns(config.MaxIdleConnections)
	}
	if config.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	}
	if config.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	}

	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		return nil, err
	}

	return db, nil
}

func getMigrationsDir(domain string) (string, error) {
	if root, _ := findModuleRoot(); root != "" {
		if p := joinAndCheck(root, domain); p != "" {
			return p, nil
		}
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		p := joinAndCheck(wd, domain)
		return p, nil
	}
	return "", fmt.Errorf("migrations directory not found")
}

func findModuleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	bi, _ := debug.ReadBuildInfo()

	for d := wd; d != "/" && d != "."; d = filepath.Dir(d) {
		gm := filepath.Join(d, "go.mod")
		if _, err := os.Stat(gm); err == nil {
			_ = bi
			return d, nil
		}
	}
	return wd, nil
}

func joinAndCheck(root, domain string) string {
	p := filepath.Join(root, "src", "main", domain, "resources", "migrations")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

// Migrate applies all SQL migrations
func Migrate(db *gorm.DB, domain string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	migrationsDir, err := getMigrationsDir(domain)
	if err != nil {
		return fmt.Errorf("failed to locate migrations: %w", err)
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []os.DirEntry
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			migrations = append(migrations, file)
		}
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name() < migrations[j].Name()
	})

	if _, err := sqlDB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	for _, migration := range migrations {
		var version string
		err := sqlDB.QueryRow(`
			SELECT version FROM schema_migrations WHERE version = $1
		`, migration.Name()).Scan(&version)

		if err == nil {
			continue
		} else if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, migration.Name()))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migration.Name(), err)
		}

		tx, err := sqlDB.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", migration.Name(), err)
		}

		if _, err := tx.Exec(`
			INSERT INTO schema_migrations (version) VALUES ($1)
		`, migration.Name()); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", migration.Name(), err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.Name(), err)
		}

		logger.Info("Applied migration", "domain", domain, "migration", migration.Name())
	}

	return nil
}
