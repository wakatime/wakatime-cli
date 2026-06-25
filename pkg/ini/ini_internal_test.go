package ini

import (
	"os"
	"path/filepath"
	"testing"
	"time"

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

	clock := &mutexClock{delay: time.Millisecond}
	assert.False(t, clock.Now().IsZero())

	select {
	case <-clock.After(time.Hour):
	case <-time.After(time.Second):
		t.Fatal("mutex clock did not fire")
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
