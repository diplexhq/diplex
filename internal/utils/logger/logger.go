// Package logger provides Logger implementations.
package logger

import (
	"log/slog"
	"os"

	"github.com/diplexhq/diplex/internal/config"
)

// New creates a configured *slog.Logger from a Config.
func New(cfg *config.Config) (Logger, error) {
	var level slog.Level

	switch {
	case cfg.Silent():
		level = slog.LevelError
	case cfg.Verbose():
		level = slog.LevelDebug
	default:
		level = slog.LevelInfo
	}

	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})), nil
}
