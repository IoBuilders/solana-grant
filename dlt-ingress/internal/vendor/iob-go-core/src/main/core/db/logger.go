package db

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm/logger"
)

type dbLogger struct {
	gormLogger      logger.Interface
	slogInnerLogger *slog.Logger
}

func newDatabaseLogger() *dbLogger {
	slogInnerLogger := slog.Default()

	gormLogger := logger.NewSlogLogger(
		slogInnerLogger,
		logger.Config{
			LogLevel: logger.Info, // The highest level to let the slog logger under the hood decide
		},
	)

	return &dbLogger{
		gormLogger:      gormLogger,
		slogInnerLogger: slogInnerLogger,
	}
}

func (l *dbLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l.gormLogger.LogMode(level)
}

func (l *dbLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.gormLogger.Info(ctx, msg, data)
}

func (l *dbLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.gormLogger.Warn(ctx, msg, data)
}

func (l *dbLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.gormLogger.Error(ctx, msg, data)
}

func (l *dbLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.slogInnerLogger.Enabled(ctx, slog.LevelDebug) {
		l.gormLogger.Trace(ctx, begin, fc, err)
	}
}
