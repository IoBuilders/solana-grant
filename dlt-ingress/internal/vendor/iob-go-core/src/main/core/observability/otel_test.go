package observability

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

// Note: This tests has more sense with an observability local stack.
func TestCore_BasicAutoExport(t *testing.T) {

	_ = os.Setenv("OTEL_SERVICE_NAME", "service-test")
	_ = os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.version=1.2.3,deployment.environment=dev,team=platform")
	_ = os.Setenv("OTEL_TRACES_EXPORTER", "none")
	_ = os.Setenv("OTEL_LOGS_EXPORTER", "none")

	ctx := context.Background()
	shutdown, err := SetupOTelSDK(ctx)

	assert.Nil(t, err)
	logger.Info("Test_ExporterLogger - Info", "key1", "value")

	tr := otel.Tracer("test-scope")
	_, span := tr.Start(ctx, "test-span")
	span.SetStatus(codes.Ok, "ok")
	span.End()

	t.Cleanup(func() {
		_ = shutdown(ctx)
		cleanEnv(t)
	})
}

func cleanEnv(t *testing.T) {

	envVars := []string{
		"OTEL_SERVICE_NAME",
		"OTEL_RESOURCE_ATTRIBUTES",
		"OTEL_TRACES_EXPORTER",
		"OTEL_LOGS_EXPORTER",
	}

	for _, envVar := range envVars {
		err := os.Unsetenv(envVar)
		if err != nil {
			t.Fatalf("Failed to unset env var: %s", envVar)
		}
	}

}
