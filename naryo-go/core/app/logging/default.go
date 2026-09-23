package logging

import (
	"context"
	"log/slog"
)

// defaultLogger is the Logger every package-level function delegates to. It
// starts backed by log/slog.Default() and is meant to be replaced wholesale
// by SetDefaultLogger, e.g. by a host application embedding this module with
// its own logger.
var defaultLogger Logger = NewSlogLogger(slog.Default())

// SetDefaultLogger replaces the Logger used by Debug, Info, Warn and Error.
func SetDefaultLogger(logger Logger) {
	defaultLogger = logger
}

// Debug logs msg at debug level through the default Logger.
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Info logs msg at info level through the default Logger.
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Warn logs msg at warn level through the default Logger.
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error logs msg at error level through the default Logger.
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// DebugWithCtx logs msg at debug level through the default Logger.
func DebugWithCtx(ctx context.Context, msg string, args ...any) {
	defaultLogger.DebugWithCtx(ctx, msg, args...)
}

// InfoWithCtx logs msg at info level through the default Logger.
func InfoWithCtx(ctx context.Context, msg string, args ...any) {
	defaultLogger.InfoWithCtx(ctx, msg, args...)
}

// WarnWithCtx logs msg at warn level through the default Logger.
func WarnWithCtx(ctx context.Context, msg string, args ...any) {
	defaultLogger.WarnWithCtx(ctx, msg, args...)
}

// ErrorWithCtx logs msg at error level through the default Logger.
func ErrorWithCtx(ctx context.Context, msg string, args ...any) {
	defaultLogger.ErrorWithCtx(ctx, msg, args...)
}
