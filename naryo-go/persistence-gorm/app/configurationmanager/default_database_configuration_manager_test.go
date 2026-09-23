//go:build test

package configurationmanager

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/domain/database"
)

// --- stubs ---

type stubDatabaseSourceProvider struct {
	item descriptor.Database
	err  error
}

func (s *stubDatabaseSourceProvider) Load(_ context.Context) (descriptor.Database, error) {
	return s.item, s.err
}

func (s *stubDatabaseSourceProvider) Priority() int { return 1 }

type stubDatabaseDescriptor struct {
	result *database.Database
	err    error
}

func (s *stubDatabaseDescriptor) Map() (*database.Database, error) {
	return s.result, s.err
}

func newDatabase(t *testing.T) *database.Database {
	t.Helper()
	db, err := database.NewDatabase("postgres://localhost/db", 0, 0, 0, 0, 0, 0, 0, 0)
	require.NoError(t, err)
	return db
}

// --- tests ---

func TestDefaultDatabaseConfigurationManager_Load_Success(t *testing.T) {
	expected := newDatabase(t)
	provider := &stubDatabaseSourceProvider{
		item: &stubDatabaseDescriptor{result: expected},
	}
	cm := NewDefaultDatabaseConfigurationManager([]sourceprovider.DatabaseSourceProvider{provider})

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestDefaultDatabaseConfigurationManager_Load_ProviderError(t *testing.T) {
	providerErr := errors.New("source unavailable")
	provider := &stubDatabaseSourceProvider{err: providerErr}
	cm := NewDefaultDatabaseConfigurationManager([]sourceprovider.DatabaseSourceProvider{provider})

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, providerErr)
}

func TestDefaultDatabaseConfigurationManager_Load_DescriptorMapError(t *testing.T) {
	mapErr := errors.New("invalid descriptor")
	provider := &stubDatabaseSourceProvider{
		item: &stubDatabaseDescriptor{err: mapErr},
	}
	cm := NewDefaultDatabaseConfigurationManager([]sourceprovider.DatabaseSourceProvider{provider})

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, mapErr)
}
