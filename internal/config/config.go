// Package config parses CLI flags and provides runtime settings for diplex.
package config

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	defaultScanDir     = "internal"
	defaultOutputDir   = "internal/generated/diplex"
	defaultSkipPattern = `(internal\/generated\/diplex|tests|mocks?|_test\.go|_mock\.go)$`
	defaultDIDirs      = "internal/di"
)

var moduleRe = regexp.MustCompile(`^module\s+(\S+)`)

// Config holds parsed CLI flags and runtime settings.
// Simple fields — narrow interfaces are declared at the point of use.
type Config struct {
	scanDirs    []string
	outputDir   string
	skipPattern *regexp.Regexp
	module      string
	silent      bool
	verbose     bool
	diDirs      []string
}

// NewConfig parses CLI flags and returns a populated Config struct.
// Calls flag.Parse() internally — must be called once at startup.
func NewConfig() *Config {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: diplex [flags]\n\n")
		fmt.Fprintf(os.Stderr, "DIplex scans Go code and generates a high-performance DI container.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fmt.Fprintf(os.Stderr, "  -scan string\n        comma-separated directories to scan (default %q)\n", defaultScanDir)
		fmt.Fprintf(os.Stderr, "  -out string\n        output directory for generated files (default %q)\n", defaultOutputDir)
		fmt.Fprintf(os.Stderr, "  -skip string\n        regexp for skipped files/dirs (default %q)\n", defaultSkipPattern)
		fmt.Fprintf(os.Stderr, "  -module string\n        module path (overrides go.mod)\n")
		fmt.Fprintf(os.Stderr, "  -v    verbose output\n")
		fmt.Fprintf(os.Stderr, "  -s    silent mode\n")
	}

	scanDirs := flag.String("scan", defaultScanDir, "comma-separated directories to scan")
	outputDir := flag.String("out", defaultOutputDir, "output directory for generated files")
	skipPattern := flag.String("skip", defaultSkipPattern, "regexp for skipped files/dirs")
	verbose := flag.Bool("v", false, "verbose output")
	silent := flag.Bool("s", false, "silent mode")
	moduleFlag := flag.String("module", "", "module path (overrides go.mod)")
	diDirs := flag.String("di", defaultDIDirs, "comma-separated directories with DI facade interfaces")

	flag.Parse()

	if *verbose && *silent {
		panic("-v and -s are mutually exclusive")
	}

	module := *moduleFlag
	if module == "" {
		module = readModule()
		if module == "" {
			panic("go.mod not found, specify -module")
		}
	}

	return &Config{
		scanDirs:    strings.Split(*scanDirs, ","),
		outputDir:   *outputDir,
		skipPattern: regexp.MustCompile(*skipPattern),
		module:      module,
		silent:      *silent,
		verbose:     *verbose,
		diDirs:      strings.Split(*diDirs, ","),
	}
}

// readModule reads the module path from go.mod file in the current directory.
// Returns empty string if go.mod not found or module line is missing.
func readModule() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return ""
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if moduleRe.MatchString(line) {
			return moduleRe.FindStringSubmatch(line)[1]
		}
	}

	return ""
}

// ScanDirs returns the comma-separated directories to scan.
func (c *Config) ScanDirs() []string { return c.scanDirs }

// OutputDir returns the output directory for generated files.
func (c *Config) OutputDir() string { return c.outputDir }

// SkipPattern returns the regexp for skipped files/dirs.
func (c *Config) SkipPattern() *regexp.Regexp { return c.skipPattern }

// Module returns the Go module path.
func (c *Config) Module() string { return c.module }

// Silent reports whether output is suppressed.
func (c *Config) Silent() bool { return c.silent }

// Verbose reports whether verbose output is enabled.
func (c *Config) Verbose() bool { return c.verbose }

// DIDirs returns directories with DI facade interfaces.
func (c *Config) DIDirs() []string { return c.diDirs }
