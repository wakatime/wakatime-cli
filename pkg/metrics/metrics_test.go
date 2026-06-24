package metrics_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/metrics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartProfiling(t *testing.T) {
	home := t.TempDir()
	t.Setenv("WAKATIME_HOME", home)

	stop, err := metrics.StartProfiling(t.Context())
	require.NoError(t, err)
	require.NotNil(t, stop)

	stop()

	entries, err := os.ReadDir(filepath.Join(home, "metrics"))
	require.NoError(t, err)

	var cpuProfiles, memProfiles int

	for _, entry := range entries {
		switch {
		case strings.HasPrefix(entry.Name(), "cpu_") && strings.HasSuffix(entry.Name(), ".profile"):
			cpuProfiles++
		case strings.HasPrefix(entry.Name(), "mem_") && strings.HasSuffix(entry.Name(), ".profile"):
			memProfiles++
		}
	}

	assert.Equal(t, 1, cpuProfiles)
	assert.Equal(t, 1, memProfiles)
}

func TestStartProfiling_MetricsFolderError(t *testing.T) {
	home := filepath.Join(t.TempDir(), "wakatime-home")
	require.NoError(t, os.WriteFile(home, []byte("not a directory"), 0600))
	t.Setenv("WAKATIME_HOME", home)

	stop, err := metrics.StartProfiling(t.Context())

	require.Error(t, err)
	assert.Nil(t, stop)
	assert.Contains(t, err.Error(), "failed to create metrics folder")
}
