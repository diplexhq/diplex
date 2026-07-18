package tests

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/diplexhq/diplex/internal/generator"
	"github.com/diplexhq/diplex/internal/parser"
	"github.com/diplexhq/diplex/internal/resolver"
	"github.com/diplexhq/diplex/internal/scanner"
	"github.com/diplexhq/diplex/internal/utils/logger"
)

type stubConfig struct {
	scanDirs    []string
	skipPattern *regexp.Regexp
	diDirs      []string
	outputDir   string
	module      string
}

func (s *stubConfig) ScanDirs() []string          { return s.scanDirs }
func (s *stubConfig) SkipPattern() *regexp.Regexp { return s.skipPattern }
func (s *stubConfig) DIDirs() []string            { return s.diDirs }
func (s *stubConfig) OutputDir() string           { return s.outputDir }
func (s *stubConfig) Module() string              { return s.module }

// expectedHash — SHA-256 hash of the generated di.go.
// To update after generator changes: `sha256sum internal/tests/generated/diplex/di.go`.
const expectedHash = "50840a4e9f82bf728c17d42040fee669a22cb894b8e7d688f27ea3eada3bbb0a"

// TestDI — full DI pipeline: scan → parse → resolve → generate + hash verification.
// If generation logic changes, the test fails — update expectedHash constant.
func TestDI(t *testing.T) {
	t.Parallel()

	orig, _ := os.Getwd()

	_ = os.Chdir("../..")
	defer func() { _ = os.Chdir(orig) }()

	cfg := &stubConfig{
		scanDirs:    []string{"internal/tests"},
		skipPattern: regexp.MustCompile(`(internal\/generated\/diplex|_test\.go|_mock\.go)$`),
		diDirs:      []string{"internal/tests/di"},
		outputDir:   "internal/tests/generated/diplex",
		module:      "github.com/diplexhq/diplex",
	}
	log := logger.Noop{}

	files := scanner.New(log, cfg).Scan()

	parsedData := parser.New(log, cfg).Parse(files)

	resolvedData := resolver.New(cfg).Resolve(parsedData)

	generator.New(log, cfg).Generate(resolvedData)

	// Verify generated di.go hash.
	generated := filepath.Join(cfg.outputDir, "di.go")

	data, err := os.ReadFile(generated)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}

	got := sha256.Sum256(data)
	if gotHex := hex.EncodeToString(got[:]); gotHex != expectedHash {
		t.Fatalf("generated di.go hash mismatch: got %s, want %s; run: sha256sum %s", gotHex, expectedHash, generated)
	}
}

func BenchmarkParse(b *testing.B) {
	orig, _ := os.Getwd()

	_ = os.Chdir("../..")
	defer func() { _ = os.Chdir(orig) }()

	cfg := &stubConfig{
		scanDirs:    []string{"internal/tests"},
		skipPattern: regexp.MustCompile(`(internal\/generated\/diplex|_test\.go|_mock\.go)$`),
		diDirs:      []string{"internal/tests/di"},
		outputDir:   "internal/tests/generated/diplex",
		module:      "github.com/diplexhq/diplex",
	}

	log := logger.Noop{}

	b.ResetTimer()
	b.ReportAllocs()

	p := parser.New(log, cfg)

	s := scanner.New(log, cfg)
	for b.Loop() {
		p.Parse(s.Scan())
	}
}

func BenchmarkResolve(b *testing.B) {
	orig, _ := os.Getwd()

	_ = os.Chdir("../..")
	defer func() { _ = os.Chdir(orig) }()

	cfg := &stubConfig{
		scanDirs:    []string{"internal/tests"},
		skipPattern: regexp.MustCompile(`(internal\/generated\/diplex|_test\.go|_mock\.go)$`),
		diDirs:      []string{"internal/tests/di"},
		outputDir:   b.TempDir(),
		module:      "github.com/diplexhq/diplex",
	}

	log := logger.Noop{}

	b.ResetTimer()
	b.ReportAllocs()

	data := parser.New(log, cfg).Parse(scanner.New(log, cfg).Scan())

	r := resolver.New(cfg)
	for b.Loop() {
		r.Resolve(data)
	}
}

func BenchmarkGenerate(b *testing.B) {
	orig, _ := os.Getwd()

	_ = os.Chdir("../..")
	defer func() { _ = os.Chdir(orig) }()

	cfg := &stubConfig{
		scanDirs:    []string{"internal/tests"},
		skipPattern: regexp.MustCompile(`(internal\/generated\/diplex|_test\.go|_mock\.go)$`),
		diDirs:      []string{"internal/tests/di"},
		outputDir:   "internal/tests/generated/diplex",
		module:      "github.com/diplexhq/diplex",
	}

	log := logger.Noop{}

	data := resolver.New(cfg).Resolve(parser.New(log, cfg).Parse(scanner.New(log, cfg).Scan()))
	g := generator.New(log, cfg)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		g.Generate(data)
	}
}

func BenchmarkTotal(b *testing.B) {
	orig, _ := os.Getwd()

	_ = os.Chdir("../..")
	defer func() { _ = os.Chdir(orig) }()

	cfg := &stubConfig{
		scanDirs:    []string{"internal/tests"},
		skipPattern: regexp.MustCompile(`(internal\/generated\/diplex|_test\.go|_mock\.go)$`),
		diDirs:      []string{"internal/tests/di"},
		outputDir:   "internal/tests/generated/diplex",
		module:      "github.com/diplexhq/diplex",
	}

	log := logger.Noop{}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		generator.New(log, cfg).Generate(resolver.New(cfg).Resolve(parser.New(log, cfg).Parse(scanner.New(log, cfg).Scan())))
	}
}
