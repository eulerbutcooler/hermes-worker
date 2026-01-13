package logger

import (
	"log/slog"
	"os"
	"time"
)

type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

type Config struct {
	Level       LogLevel
	ServiceName string
	Environment string
	Pretty      bool
}

func New(cfg Config) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.Level == LevelDebug,
	}

	if cfg.Pretty {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler).With(
		slog.String("service", cfg.ServiceName),
		slog.String("environment", cfg.Environment),
		slog.String("version", "1.0.0"),
	)

	slog.SetDefault(logger)
	return logger
}

// func WithContext(ctx context.Context, logger *slog.Logger) *slog.Logger {
// 	return logger
// }

func LogDuration(logger *slog.Logger, operation string, start time.Time) {
	duration := time.Since(start)
	logger.Info("operation completed",
		slog.String("operation", operation),
		slog.Duration("duration", duration))
}
