package logger

import (
	"context"
)

var defaultLogger Logger

func SetDefaultLogger(logger Logger) {
	defaultLogger = logger
}

type Logger interface {
	Debug(message string, args ...any)
	Info(message string, args ...any)
	Warn(message string, args ...any)
	Error(message string, args ...any)
	DebugWithCtx(ctx context.Context, message string, args ...any)
	InfoWithCtx(ctx context.Context, message string, args ...any)
	WarnWithCtx(ctx context.Context, message string, args ...any)
	ErrorWithCtx(ctx context.Context, message string, args ...any)
}

func Debug(message string, args ...any) {
	defaultLogger.Debug(message, args...)
}
func Info(message string, args ...any) {
	defaultLogger.Info(message, args...)
}
func Warn(message string, args ...any) {
	defaultLogger.Warn(message, args...)
}
func Error(message string, args ...any) {
	defaultLogger.Error(message, args...)
}
func DebugWithCtx(ctx context.Context, message string, args ...any) {
	defaultLogger.DebugWithCtx(ctx, message, args...)
}
func InfoWithCtx(ctx context.Context, message string, args ...any) {
	defaultLogger.InfoWithCtx(ctx, message, args...)
}
func WarnWithCtx(ctx context.Context, message string, args ...any) {
	defaultLogger.WarnWithCtx(ctx, message, args...)
}
func ErrorWithCtx(ctx context.Context, message string, args ...any) {
	defaultLogger.ErrorWithCtx(ctx, message, args...)
}
