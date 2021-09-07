package heartbeat

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsXCodePlayground(t *testing.T) {
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
			defer os.RemoveAll(test.Dir)

			ret := isXCodePlayground(test.Dir)

			assert.Equal(t, test.Expected, ret)
		})
	}
}

func setupTestXCodePlayground(t *testing.T, dir string) string {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime")
	require.NoError(t, err)

	err = os.Mkdir(filepath.Join(tmpDir, dir), os.FileMode(int(0700)))
	require.NoError(t, err)

	return filepath.Join(tmpDir, dir)
}
