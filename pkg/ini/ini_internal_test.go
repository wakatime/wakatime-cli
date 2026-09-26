package ini

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/juju/mutex"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInternalHelpers(t *testing.T) {
	assert.Equal(t, "settings.api_key", sectionKey("settings", "api_key"))
	assert.Equal(t, "api_key", sectionKey("DEFAULT", "api_key"))

	sanitized := string(sanitizeMalformedSections([]byte("[good]\na=1\n[bad\nb=2\n[other]\nc=3")))
	assert.Contains(t, sanitized, "[good]")
	assert.NotContains(t, sanitized, "[bad")
	assert.Contains(t, sanitized, "[other]")

	missing := filepath.Join(t.TempDir(), "missing.cfg")
	assert.False(t, fileExists(missing))
	require.NoError(t, os.WriteFile(missing, []byte("x"), 0600))
	assert.True(t, fileExists(missing))

	clock := &mutexClock{}
	assert.False(t, clock.Now().IsZero())

	select {
	case <-clock.After(time.Millisecond):
	case <-time.After(time.Second):
		t.Fatal("mutex clock did not fire")
	}

	select {
	case <-clock.After(time.Hour):
		t.Fatal("mutex clock ignored the requested duration")
	case <-time.After(10 * time.Millisecond):
	}
}

func TestSalvageConfigBranches(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "wakatime.cfg")
	require.NoError(t, os.WriteFile(tmp, []byte("[settings]\napi_key = value\x00\n[broken\nignored = true\n"), 0600))

	v := viper.New()
	require.NoError(t, salvageConfig(v, tmp))
	assert.Equal(t, "value", v.GetString("settings.api_key"))
	assert.Empty(t, v.GetString("broken.ignored"))

	err := salvageConfig(v, filepath.Join(t.TempDir(), "missing.cfg"))
	require.Error(t, err)
}

func TestWriteDoesNotProceedWithoutLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "internal.cfg")
	original := []byte("[internal]\nexisting = keep\n")
	require.NoError(t, os.WriteFile(path, original, 0o600))
	writer, err := NewWriter(t.Context(), viper.New(), func(context.Context, *viper.Viper) (string, error) {
		return path, nil
	})
	require.NoError(t, err)
	lock, err := mutex.Acquire(mutex.Spec{Name: "wakatime-cli-config-mutex", Delay: time.Millisecond,
		Timeout: time.Second, Clock: &mutexClock{}})
	require.NoError(t, err)

	defer lock.Release()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	require.ErrorContains(t, writer.Write(ctx, "internal", map[string]string{"new": "value"}), "config mutex")

	actual, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, original, actual)
}
