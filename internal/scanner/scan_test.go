package scanner_test

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"testing"

	. "github.com/diplexhq/diplex/internal/scanner"

	"github.com/diplexhq/diplex/internal/domain"
	"github.com/diplexhq/diplex/internal/utils/logger"
)

type stubScanConfig struct {
	cwd         string
	scanDirs    []string
	skipPattern *regexp.Regexp
}

func (s *stubScanConfig) Cwd() string                 { return s.cwd }
func (s *stubScanConfig) ScanDirs() []string          { return s.scanDirs }
func (s *stubScanConfig) SkipPattern() *regexp.Regexp { return s.skipPattern }

// collectFiles drains the iterator into a slice.
func collectFiles(files domain.SourceFiles) []domain.SourceFile {
	var result []domain.SourceFile
	for f := range files {
		result = append(result, f)
	}

	return result
}

// touch creates a file inside dir with empty content.
func touch(t *testing.T, dir, relPath string) {
	t.Helper()

	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
	}

	if err := os.WriteFile(full, []byte{}, 0o600); err != nil {
		t.Fatalf("touch %s: %v", full, err)
	}
}

func TestScan_Combined(t *testing.T) {
	dir := t.TempDir()

	touch(t, dir, "a.go")
	touch(t, dir, "b.go")
	touch(t, dir, "sub/c.go")
	touch(t, dir, "keep.go")
	touch(t, dir, "keep2.go")
	touch(t, dir, "root.go")
	touch(t, dir, "internal/svc.go")
	touch(t, dir, "main.go")
	touch(t, dir, "alpha/alpha.go")
	touch(t, dir, "beta/beta.go")

	// Skipped by skipPattern: vendor/ directory
	touch(t, dir, "vendor/lib.go")

	// Skipped by skipPattern: matches _test.go suffix
	touch(t, dir, "skip_test.go")

	// Skipped by skipPattern: matches mock.go suffix
	touch(t, dir, "some_mock.go")

	// Skipped by skipPattern: matches tests directory
	touch(t, dir, "tests/test_file.go")

	// Skipped: not a .go file (scanner filter)
	touch(t, dir, "README.md")
	touch(t, dir, "Makefile")
	touch(t, dir, "config.yaml")

	scanner := New(logger.Noop{}, &stubScanConfig{
		cwd:         dir,
		scanDirs:    []string{dir},
		skipPattern: regexp.MustCompile(`((internal\/generated\/diplex|tests|mocks?|_test\.go|_mock\.go|vendor)$)`),
	})

	files := collectFiles(scanner.Scan())

	expected := []domain.SourceFile{
		domain.SourceFile(dir + "/a.go"),
		domain.SourceFile(dir + "/alpha/alpha.go"),
		domain.SourceFile(dir + "/b.go"),
		domain.SourceFile(dir + "/beta/beta.go"),
		domain.SourceFile(dir + "/internal/svc.go"),
		domain.SourceFile(dir + "/keep.go"),
		domain.SourceFile(dir + "/keep2.go"),
		domain.SourceFile(dir + "/main.go"),
		domain.SourceFile(dir + "/root.go"),
		domain.SourceFile(dir + "/sub/c.go"),
	}

	sort.Slice(files, func(i, j int) bool {
		return string(files[i]) < string(files[j])
	})

	if !reflect.DeepEqual(files, expected) {
		t.Errorf("files mismatch:\ngot:      %v\nexpected: %v", files, expected)
	}
}
