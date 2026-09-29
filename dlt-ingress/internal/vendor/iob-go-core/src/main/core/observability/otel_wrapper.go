package observability

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

// OtelLogLevelHandler wraps otelslog.Handler with a configurable log level,
// since the default OTel slog bridge does not support level filtering.
type OtelLogLevelHandler struct {
	*otelslog.Handler
	Level slog.Level
}

func (h *OtelLogLevelHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.Level
}

func NewOtelLogLevelHandler(level slog.Level, scope Scope, options ...otelslog.Option) *OtelLogLevelHandler {
	return &OtelLogLevelHandler{
		Handler: otelslog.NewHandler(scope.ScopeName, options...),
		Level:   level,
	}
}
