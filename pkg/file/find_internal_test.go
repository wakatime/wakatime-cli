package file

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRootPath(t *testing.T) {
	tests := map[string]string{
		"empty":                "",
		"dot":                  ".",
		"file path separattor": string(filepath.Separator),
		"parent directory":     filepath.Dir("/"),
		"volume name":          filepath.VolumeName("some/path/to/file.go"),
		"wsl root":             "\\\\wsl$",
		"drive letter":         "C:\\",
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ok := isRootPath(test)
			assert.True(t, ok)
		})
	}
}
