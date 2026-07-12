package scanner

import "regexp"

// Config is what Scanner needs from the application config.
// Declared locally — each consumer defines exactly what it needs.
type Config interface {
	ScanDirs() []string
	SkipPattern() *regexp.Regexp
}
