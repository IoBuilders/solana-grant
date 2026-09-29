package logger

import (
	"context"
	"log/slog"
)

type SlogLogger struct {
	l *slog.Logger
}

func NewSlogLogger(logger *slog.Logger) Logger {
	return &SlogLogger{
		l: logger,
	}
}

func (z *SlogLogger) Debug(message string, args ...any) {
	z.l.Debug(message, args...)
}
func (z *SlogLogger) Info(message string, args ...any) {
	z.l.Info(message, args...)
}
func (z *SlogLogger) Warn(message string, args ...any) {
	z.l.Warn(message, args...)
}
func (z *SlogLogger) Error(message string, args ...any) {
	z.l.Error(message, args...)
}
func (z *SlogLogger) DebugWithCtx(ctx context.Context, message string, args ...any) {
	z.l.DebugContext(ctx, message, args...)
}
func (z *SlogLogger) InfoWithCtx(ctx context.Context, message string, args ...any) {
	z.l.InfoContext(ctx, message, args...)
}
func (z *SlogLogger) WarnWithCtx(ctx context.Context, message string, args ...any) {
	z.l.WarnContext(ctx, message, args...)
}
func (z *SlogLogger) ErrorWithCtx(ctx context.Context, message string, args ...any) {
	z.l.ErrorContext(ctx, message, args...)
}
