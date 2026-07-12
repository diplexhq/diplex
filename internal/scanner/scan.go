package scanner

import (
	iofs "io/fs"
	"path/filepath"
	"strings"

	"github.com/diplexhq/diplex/internal/domain"
	"github.com/diplexhq/diplex/internal/utils"
)

// Scan returns a channel of Go source files.
// Uses Cwd, ScanDirs, and SkipPattern from the injected Config.
//
// Note: To scan internal/di directory, add it explicitly to -scan flag
// (e.g., "-scan=.,internal/di"). The DIDirs config is used only for
// filtering which interfaces should get DI facades.
func (fs *Scanner) Scan() domain.SourceFiles {
	ch := make(domain.SourceFiles, 4)

	go func() {
		defer close(ch)

		for _, dir := range fs.scanDirs {
			fs.scanDir(dir, ch)
		}
	}()

	return ch
}

// scanDir walks the directory and sends files to the channel.
func (fs *Scanner) scanDir(dir string, ch domain.SourceFiles) {
	fs.logger.Debug("scanning directory", "dir", dir)

	utils.NoErr(filepath.WalkDir(dir, func(path string, d iofs.DirEntry, err error) error {
		utils.NoErr(err)

		if d.IsDir() {
			if fs.skipPattern.MatchString(path) {
				fs.logger.Debug("skipping directory", "name", d.Name())
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		if fs.skipPattern.MatchString(path) {
			fs.logger.Debug("skipping path", "name", d.Name())
			return nil
		}

		ch <- domain.SourceFile(path)

		return nil
	}))
}
