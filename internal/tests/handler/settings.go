package handler

import (
	"github.com/diplexhq/diplex/internal/tests/config"
	"github.com/diplexhq/diplex/internal/tests/logger"
)

// Settings — exported fields are DI dependencies.
type Settings struct {
	Config *config.Config
	Logger logger.Logger
}

func NewSettings(cfg *config.Config, log logger.Logger) *Settings {
	return &Settings{Config: cfg, Logger: log}
}
