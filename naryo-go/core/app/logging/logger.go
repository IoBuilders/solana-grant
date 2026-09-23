package logging

import "context"

// Logger is the application port for structured, leveled logging. Adapters in
// the infrastructure layer back it with a concrete logger (e.g. log/slog). The
// variadic args are key/value pairs, matching the log/slog convention.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	DebugWithCtx(ctx context.Context, msg string, args ...any)
	InfoWithCtx(ctx context.Context, msg string, args ...any)
	WarnWithCtx(ctx context.Context, msg string, args ...any)
	ErrorWithCtx(ctx context.Context, msg string, args ...any)
}
