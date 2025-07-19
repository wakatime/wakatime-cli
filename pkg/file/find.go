package file

import (
	"context"
	"os"
	"path/filepath"
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/log"
)

// maxRecursiveIteration limits the number of a func will be called recursively.
const maxRecursiveIteration = 500

var driveLetterRegex = regexp.MustCompile(`^[a-zA-Z]:\\$`)

// Find searches for a file named `filename`.
// Search starts in `directory` and will traverse through all parent directories.
func Find(ctx context.Context, directory, filename string) (string, bool) {
	if directory == "" || filename == "" {
		return "", false
	}

	i := 0
	for i < maxRecursiveIteration {
		if isRootPath(directory) {
			return "", false
		}

		if Exists(filepath.Join(directory, filename)) {
			return filepath.Join(directory, filename), true
		}

		directory = filepath.Clean(filepath.Join(directory, ".."))

		i++
	}

	logger := log.Extract(ctx)
	logger.Warnf("max %d iterations reached without finding %s", maxRecursiveIteration, filename)

	return "", false
}

// Exists checks if a file exist.
func Exists(fp string) bool {
	info, err := os.Stat(fp)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func isRootPath(directory string) bool {
	return (directory == "" ||
		directory == "." ||
		directory == string(filepath.Separator) ||
		directory == filepath.Dir(directory)) ||
		directory == filepath.VolumeName(directory) ||
		directory == "\\\\wsl$" ||
		driveLetterRegex.MatchString(directory)
}
