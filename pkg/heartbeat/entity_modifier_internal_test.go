package heartbeat

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsXCodePlayground(t *testing.T) {
	ctx := context.Background()

	tests := map[string]struct {
		Dir      string
		Expected bool
	}{
		"playground directory": {
			Dir:      setupTestXCodePlayground(t, "wakatime.playground"),
			Expected: true,
		},
		"xcplayground directory": {
			Dir:      setupTestXCodePlayground(t, "wakatime.xcplayground"),
			Expected: true,
		},
		"xcplaygroundpage directory": {
			Dir:      setupTestXCodePlayground(t, "wakatime.xcplaygroundpage"),
			Expected: true,
		},
		"not playground": {
			Dir:      setupTestXCodePlayground(t, "wakatime"),
			Expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ret := isXCodePlayground(ctx, test.Dir)

			assert.Equal(t, test.Expected, ret)
		})
	}
}

func TestIsXCodeProject(t *testing.T) {
	ctx := context.Background()

	tests := map[string]struct {
		Dir      string
		Expected bool
	}{
		"project directory": {
			Dir:      setupTestXCodePlayground(t, "wakatime.xcodeproj"),
			Expected: true,
		},
		"not project": {
			Dir:      setupTestXCodePlayground(t, "wakatime"),
			Expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ret := isXCodeProject(ctx, test.Dir)

			assert.Equal(t, test.Expected, ret)
		})
	}
}

func setupTestXCodePlayground(t *testing.T, dir string) string {
	tmpDir := t.TempDir()

	err := os.Mkdir(filepath.Join(tmpDir, dir), os.FileMode(int(0700)))
	require.NoError(t, err)

	return filepath.Join(tmpDir, dir)
}
