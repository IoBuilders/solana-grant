package logger

import (
	"log/slog"
	"os"
)

// Configures default logger with default info level
func init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slogLogger := slog.New(handler)
	SetDefaultLogger(NewSlogLogger(slogLogger))
}
