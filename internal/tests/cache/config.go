package cache

import (
	"github.com/diplexhq/diplex/internal/tests/config"
	"github.com/diplexhq/diplex/internal/tests/logger"
)

// CacheConfig — auto-provider test: no explicit New function, just exported fields.
type CacheConfig struct {
	Config *config.Config
	Logger logger.Logger
}
