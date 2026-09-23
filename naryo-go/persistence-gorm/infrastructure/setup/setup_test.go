//go:build test

package persistencegormsetup

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// resetFlags lets coreconfig.LoadConfig re-register its "config" flag on
// each test call without panicking: LoadConfig calls flag.String("config",
// ...) on every invocation, and registering the same flag twice on the
// process-global flag.CommandLine panics.
func resetFlags(t *testing.T) {
	t.Helper()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestSetup_LoadConfigError_RootPropertyNotFound(t *testing.T) {
	resetFlags(t)
	t.Setenv("CONFIG_PATH", "")

	module, err := Setup(context.Background(), "naryo:\n  database:\n    url: postgres://localhost/db\n", "nonexistent")

	assert.ErrorContains(t, err, "root property")
	assert.Nil(t, module)
}

// TestSetup_MissingDatabaseUrl_ReturnsValidationErrorWithoutConnecting covers
// the one Setup failure reachable without a real Postgres server: an invalid
// database config fails dbConfigManager.Load before newPostgresDB ever runs,
// so no connection is attempted.
func TestSetup_MissingDatabaseUrl_ReturnsValidationErrorWithoutConnecting(t *testing.T) {
	resetFlags(t)
	t.Setenv("CONFIG_PATH", "")

	yaml := "naryo:\n  database:\n    maxOpenConnections: 10\n"

	module, err := Setup(context.Background(), yaml, "naryo")

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
	assert.Nil(t, module)
}
