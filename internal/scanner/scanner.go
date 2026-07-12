package scanner

import (
	"regexp"

	"github.com/diplexhq/diplex/internal/utils/logger"
)

// Scanner scans Go files.
type Scanner struct {
	logger      logger.Logger
	scanDirs    []string
	skipPattern *regexp.Regexp
}

// New creates a Scanner with the given logger and config.
func New(logger logger.Logger, cfg Config) *Scanner {
	return &Scanner{
		logger:      logger,
		scanDirs:    cfg.ScanDirs(),
		skipPattern: cfg.SkipPattern(),
	}
}
