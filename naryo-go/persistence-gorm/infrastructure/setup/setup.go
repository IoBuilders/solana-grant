package persistencegormsetup

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"ariga.io/atlas/atlasexec"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	coreconfig "gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/config"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/app/configurationmanager"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/domain/database"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/infrastructure/config"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/infrastructure/eventstore"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/resources/migrations"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PersistenceGormModule struct {
	EventStores       map[reflect.Type]any
	LatestBlockStores map[reflect.Type]any
}

func Setup(ctx context.Context, applicationYml string, rootProperty string) (*PersistenceGormModule, error) {
	db, err := loadDatabaseConfig(ctx, applicationYml, rootProperty)
	if err != nil {
		return nil, err
	}
	gormDb, err := newPostgresDB(db)
	if err != nil {
		return nil, err
	}

	eventStores := make(map[reflect.Type]any)
	eventStores[reflect.TypeFor[event.SolanaBlockEvent]()] = eventstore.NewSolanaBlockEventStore(gormDb)
	eventStores[reflect.TypeFor[event.SolanaTransactionEvent]()] = eventstore.NewSolanaTransactionEventStore(gormDb)
	eventStores[reflect.TypeFor[event.SolanaContractEvent]()] = eventstore.NewSolanaContractEventStore(gormDb)

	latestBlockStores := make(map[reflect.Type]any)
	latestBlockStores[reflect.TypeFor[event.SolanaBlockEvent]()] = eventstore.NewSolanaLatestBlockStore(gormDb)

	return &PersistenceGormModule{
		EventStores:       eventStores,
		LatestBlockStores: latestBlockStores,
	}, nil
}

// Migrate runs this module's Atlas migrations (embedded from resources/migrations) against the
// configured naryo.database.url via the atlas CLI, so the schema is up to date before Setup opens
// any store. Callers (server/cmd/server/main.go) are expected to call this before Setup. Requires
// the "atlas" binary to be on PATH.
func Migrate(ctx context.Context, applicationYml string, rootProperty string) error {
	db, err := loadDatabaseConfig(ctx, applicationYml, rootProperty)
	if err != nil {
		return err
	}

	workdir, err := atlasexec.NewWorkingDir(atlasexec.WithMigrations(migrations.FS))
	if err != nil {
		return fmt.Errorf("failed to load atlas migrations: %w", err)
	}
	defer func() { _ = workdir.Close() }()

	client, err := atlasexec.NewClient(workdir.Path(), "atlas")
	if err != nil {
		return fmt.Errorf("failed to initialize atlas client: %w", err)
	}

	if _, err := client.MigrateApply(ctx, &atlasexec.MigrateApplyParams{URL: db.Url}); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

func loadDatabaseConfig(ctx context.Context, applicationYml string, rootProperty string) (*database.Database, error) {
	envProperties, err := coreconfig.LoadConfig[config.EnvironmentProperties](ctx, applicationYml, rootProperty)
	if err != nil {
		return nil, err
	}

	envDbSourceProvider := config.NewEnvDatabaseSourceProvider(envProperties)
	dbConfigManager := configurationmanager.NewDefaultDatabaseConfigurationManager([]sourceprovider.DatabaseSourceProvider{envDbSourceProvider})

	return dbConfigManager.Load(ctx)
}

func newPostgresDB(db *database.Database) (*gorm.DB, error) {
	// Add timeouts to DSN if they are not already present
	var params []string
	if !strings.Contains(db.Url, "connect_timeout=") && db.ConnectTimeout > 0 {
		// connect_timeout is in seconds
		params = append(params, fmt.Sprintf("connect_timeout=%d", int(db.ConnectTimeout.Seconds())))
	}
	if !strings.Contains(db.Url, "statement_timeout=") && db.StatementTimeout > 0 {
		// statement_timeout is in milliseconds
		params = append(params, fmt.Sprintf("statement_timeout=%d", db.StatementTimeout.Milliseconds()))
	}
	if !strings.Contains(db.Url, "idle_in_transaction_session_timeout=") && db.IdleInTransactionSessionTimeout > 0 {
		// idle_in_transaction_session_timeout is in milliseconds
		params = append(params, fmt.Sprintf("idle_in_transaction_session_timeout=%d", db.IdleInTransactionSessionTimeout.Milliseconds()))
	}
	if !strings.Contains(db.Url, "lock_timeout=") && db.LockTimeout > 0 {
		// lock_timeout is in milliseconds
		params = append(params, fmt.Sprintf("lock_timeout=%d", db.LockTimeout.Milliseconds()))
	}

	if len(params) > 0 {
		separator := "?"
		if strings.Contains(db.Url, "?") {
			separator = "&"
		}
		db.Url = fmt.Sprintf("%s%s%s", db.Url, separator, strings.Join(params, "&"))
	}

	gormDb, err := gorm.Open(postgres.New(postgres.Config{
		DSN: db.Url,
	}), &gorm.Config{
		Logger: newDatabaseLogger(),
		// Without this, GORM sends a save's cascaded associations (e.g. a block's Transactions,
		// each with their own Instructions/Accounts/Logs) as one INSERT per table, however many
		// rows there are. Postgres's extended protocol caps a statement at 65535 bind
		// parameters, which a single busy block can exceed on its own; CreateBatchSize makes
		// GORM split into chunks of this size instead. 1000 is GORM's own documented example
		// value, and stays well under the limit even for this schema's widest table (~11 columns).
		CreateBatchSize: 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	sqlDB, err := gormDb.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Set connection pool settings
	if db.MaxOpenConnections > 0 {
		sqlDB.SetMaxOpenConns(db.MaxOpenConnections)
	}
	if db.MaxIdleConnections > 0 {
		sqlDB.SetMaxIdleConns(db.MaxIdleConnections)
	}
	if db.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(db.ConnMaxLifetime)
	}
	if db.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(db.ConnMaxIdleTime)
	}

	return gormDb, nil
}
