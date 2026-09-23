package logging

import (
	"context"
	"log/slog"
)

// slogLogger adapts a *slog.Logger to Logger.
type slogLogger struct {
	l *slog.Logger
}

// NewSlogLogger builds a Logger backed by l.
func NewSlogLogger(l *slog.Logger) Logger {
	return &slogLogger{l: l}
}

func (s *slogLogger) Debug(msg string, args ...any) { s.l.Debug(msg, args...) }
func (s *slogLogger) Info(msg string, args ...any)  { s.l.Info(msg, args...) }
func (s *slogLogger) Warn(msg string, args ...any)  { s.l.Warn(msg, args...) }
func (s *slogLogger) Error(msg string, args ...any) { s.l.Error(msg, args...) }
func (s *slogLogger) DebugWithCtx(ctx context.Context, msg string, args ...any) {
	s.l.DebugContext(ctx, msg, args...)
}
func (s *slogLogger) InfoWithCtx(ctx context.Context, msg string, args ...any) {
	s.l.InfoContext(ctx, msg, args...)
}
func (s *slogLogger) WarnWithCtx(ctx context.Context, msg string, args ...any) {
	s.l.WarnContext(ctx, msg, args...)
}
func (s *slogLogger) ErrorWithCtx(ctx context.Context, msg string, args ...any) {
	s.l.ErrorContext(ctx, msg, args...)
}
