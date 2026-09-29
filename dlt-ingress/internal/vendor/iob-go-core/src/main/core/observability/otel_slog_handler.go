package observability

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// OtelSlogTextHandler is a text slog.Handler that injects otel traces attributes traceId and spanId from context
type OtelSlogTextHandler struct {
	*slog.TextHandler
}

func NewOtelSlogTextHandler(level slog.Level) *OtelSlogTextHandler {
	return &OtelSlogTextHandler{
		TextHandler: slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		}),
	}
}

func (h *OtelSlogTextHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx == nil {
		return h.TextHandler.Handle(ctx, r)
	}
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		r.AddAttrs(
			slog.String("traceId", spanContext.TraceID().String()),
			slog.String("spanId", spanContext.SpanID().String()),
		)
	}

	return h.TextHandler.Handle(ctx, r)
}
