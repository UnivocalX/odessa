package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Config holds the connection settings for the database.
type Config struct {
	// DSN is the PostgreSQL connection string.
	// Example: "postgres://user:pass@localhost/odessa?sslmode=disable"
	//          "host=localhost port=5432 dbname=odessa user=pg password=secret sslmode=disable"
	DSN Secret
}

// Open opens a GORM PostgreSQL database using cfg.
func Open(cfg Config) (*gorm.DB, error) {
	slog.Info("opening database", "driver", "postgres")

	db, err := gorm.Open(postgres.Open(cfg.DSN.Expose()), &gorm.Config{
		Logger: newSlogLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("repository: open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("repository: get sql.DB: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("repository: ping postgres: %w", err)
	}

	slog.Info("database connected", "driver", "postgres")
	return db, nil
}

// Close closes the underlying connection pool.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	slog.Info("closing database")
	return sqlDB.Close()
}

// slogLogger adapts slog to the gorm logger.Interface.
type slogLogger struct{}

func newSlogLogger() gormlogger.Interface {
	return &slogLogger{}
}

func (l *slogLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *slogLogger) Info(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, fmt.Sprintf(msg, args...))
}

func (l *slogLogger) Warn(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, fmt.Sprintf(msg, args...))
}

func (l *slogLogger) Error(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, fmt.Sprintf(msg, args...))
}

func (l *slogLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		slog.ErrorContext(ctx, "db query", "sql", sql, "rows", rows, "elapsed", elapsed, "error", err)
	case elapsed > 200*time.Millisecond:
		slog.WarnContext(ctx, "slow query", "sql", sql, "rows", rows, "elapsed", elapsed)
	default:
		slog.DebugContext(ctx, "db query", "sql", sql, "rows", rows, "elapsed", elapsed)
	}
}
