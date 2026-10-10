package heartbeat

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	heartbeatpkg "github.com/wakatime/wakatime-cli/pkg/heartbeat"
	offlinepkg "github.com/wakatime/wakatime-cli/pkg/offline"
	paramspkg "github.com/wakatime/wakatime-cli/pkg/params"
	"github.com/wakatime/wakatime-cli/pkg/regex"
)

func TestLoadParamsInternalBranches(t *testing.T) {
	_, err := loadParams(t.Context(), nil, paramspkg.FlagReadOrderFlagPrecedence)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "viper instance unset")

	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")

	_, err = loadParams(t.Context(), v, paramspkg.FlagReadOrderFlagPrecedence)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load heartbeat params")

	_, err = loadAISyncParams(t.Context(), viper.New(), paramspkg.FlagReadOrderFlagPrecedence)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load API params")
}

func TestApplyAIParsingDisabledPassesThrough(t *testing.T) {
	input := []heartbeatpkg.Heartbeat{{
		Entity:     "Terminal",
		EntityType: heartbeatpkg.AppType,
		Time:       1,
		UserAgent:  "plugin/0.0.1",
	}}

	got, err := applyAIParsing(t.Context(), viper.New(), paramspkg.Params{
		AI:  paramspkg.AIParams{SyncDisabled: true},
		API: paramspkg.API{Plugin: "plugin/0.0.1"},
	}, input)

	require.NoError(t, err)
	assert.Equal(t, input, got)
}

func TestRunAISyncActivityNoActivity(t *testing.T) {
	paramspkg.Once = sync.Once{}

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("offline-queue-file", "~missing-user/offline.bdb")

	code, err := RunAISyncActivity(t.Context(), v)

	require.NoError(t, err)
	assert.Equal(t, 0, code)
}

func TestRunAISyncActivityLoadParamsError(t *testing.T) {
	paramspkg.Once = sync.Once{}

	t.Setenv("WAKATIME_API_KEY", "")

	code, err := RunAISyncActivity(t.Context(), viper.New())

	require.Error(t, err)
	assert.Equal(t, 104, code)
	assert.Contains(t, err.Error(), "failed to load command params")
}

func TestSendPreparedHeartbeatsRateLimitedSavesOffline(t *testing.T) {
	queueFile := filepath.Join(t.TempDir(), "offline.bdb")
	params := internalCommandParams()
	params.Offline.LastSentAt = time.Now()
	params.Offline.RateLimit = time.Hour

	err := sendPreparedHeartbeats(
		t.Context(),
		viper.New(),
		params,
		queueFile,
		[]heartbeatpkg.Heartbeat{internalAppHeartbeat()},
		false,
		loadParams,
	)

	require.NoError(t, err)

	count, err := offlinepkg.CountHeartbeats(t.Context(), queueFile)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSendPreparedHeartbeatsBuildHandleErrorSavesOffline(t *testing.T) {
	queueFile := filepath.Join(t.TempDir(), "offline.bdb")
	params := internalCommandParams()
	params.API.SSLCertFilepath = filepath.Join(t.TempDir(), "missing.pem")

	err := sendPreparedHeartbeats(
		t.Context(),
		viper.New(),
		params,
		queueFile,
		[]heartbeatpkg.Heartbeat{internalAppHeartbeat()},
		false,
		loadParams,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize api client")

	count, countErr := offlinepkg.CountHeartbeats(t.Context(), queueFile)
	require.NoError(t, countErr)
	assert.Equal(t, 1, count)
}

func TestSendPreparedHeartbeatsSavesExtraBeforeBuildHandleError(t *testing.T) {
	queueFile := filepath.Join(t.TempDir(), "offline.bdb")
	params := internalCommandParams()
	params.API.SSLCertFilepath = filepath.Join(t.TempDir(), "missing.pem")

	heartbeats := make([]heartbeatpkg.Heartbeat, offlinepkg.SendLimit+1)
	for i := range heartbeats {
		heartbeats[i] = internalAppHeartbeat()
		heartbeats[i].Time = float64(i + 1)
	}

	err := sendPreparedHeartbeats(
		t.Context(),
		viper.New(),
		params,
		queueFile,
		heartbeats,
		false,
		loadParams,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize api client")

	count, countErr := offlinepkg.CountHeartbeats(t.Context(), queueFile)
	require.NoError(t, countErr)
	assert.Equal(t, offlinepkg.SendLimit+1, count)
}

func TestShouldUseProjectConfigBranches(t *testing.T) {
	assert.False(t, shouldUseProjectConfig([]heartbeatpkg.Heartbeat{{Entity: "Terminal", EntityType: heartbeatpkg.AppType}}))
	assert.False(t, shouldUseProjectConfig([]heartbeatpkg.Heartbeat{{
		Entity:          "/tmp/unsaved.go",
		EntityType:      heartbeatpkg.FileType,
		IsUnsavedEntity: true,
	}}))
	assert.True(t, shouldUseProjectConfig([]heartbeatpkg.Heartbeat{{
		Entity:     "/tmp/main.go",
		EntityType: heartbeatpkg.FileType,
	}}))
}

func internalCommandParams() paramspkg.Params {
	return paramspkg.Params{
		API: paramspkg.API{
			Plugin:  "plugin/0.0.1",
			Timeout: time.Second,
			URL:     "http://127.0.0.1",
		},
		Heartbeat: paramspkg.Heartbeat{
			Category:   heartbeatpkg.CodingCategory,
			Entity:     "Terminal",
			EntityType: heartbeatpkg.AppType,
			Time:       1,
		},
	}
}

func internalAppHeartbeat() heartbeatpkg.Heartbeat {
	return heartbeatpkg.Heartbeat{
		Entity:     "Terminal",
		EntityType: heartbeatpkg.AppType,
		Time:       1,
		UserAgent:  "plugin/0.0.1",
	}
}

func TestSplitAIExcludedHeartbeats_PreservesExcludedHuman(t *testing.T) {
	human := heartbeatpkg.Heartbeat{
		Entity:           "/tmp/nautilus/main.py",
		EntityType:       heartbeatpkg.FileType,
		Category:         heartbeatpkg.CodingCategory.String(),
		HumanLineChanges: heartbeatpkg.PointerTo(5),
		ProjectOverride:  "nautilus",
		Time:             1750000000,
	}

	p := internalCommandParams()
	p.AI.ExcludeProjects = []regex.Regex{regex.MustCompile("(?i)^nautilus$")} // nolint

	included, excluded := splitAIExcludedHeartbeats(t.Context(), p, []heartbeatpkg.Heartbeat{human})

	assert.Len(t, excluded, 1)
	assert.Len(t, included, 0)
	require.NotNil(t, excluded[0].Project)
	assert.Equal(t, "nautilus", *excluded[0].Project)
}

func TestSplitAIExcludedHeartbeats_KeepsNonExcluded(t *testing.T) {
	human := heartbeatpkg.Heartbeat{
		Entity:           "/tmp/other/main.py",
		EntityType:       heartbeatpkg.FileType,
		Category:         heartbeatpkg.CodingCategory.String(),
		HumanLineChanges: heartbeatpkg.PointerTo(5),
		ProjectOverride:  "other-project",
		Time:             1750000000,
	}

	p := internalCommandParams()
	p.AI.ExcludeProjects = []regex.Regex{regex.MustCompile("(?i)^nautilus$")} // nolint

	included, excluded := splitAIExcludedHeartbeats(t.Context(), p, []heartbeatpkg.Heartbeat{human})

	assert.Len(t, included, 1)
	assert.Len(t, excluded, 0)
}

func TestSplitAIExcludedHeartbeats_ProjectFileDisablesAI(t *testing.T) {
	tmpDir := t.TempDir()
	codeFile := filepath.Join(tmpDir, "main.py")
	require.NoError(t, os.WriteFile(codeFile, []byte("print('x')\n"), 0644))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, ".wakatime"),
		[]byte("[settings]\nsync_ai_disabled = true\n"),
		0644,
	))

	h := heartbeatpkg.Heartbeat{
		Entity:     codeFile,
		EntityType: heartbeatpkg.FileType,
		Category:   heartbeatpkg.CodingCategory.String(),
		Time:       1750000000,
	}

	assert.True(t, projectFileDisablesAI(t.Context(), h))

	included, excluded := splitAIExcludedHeartbeats(t.Context(), internalCommandParams(), []heartbeatpkg.Heartbeat{h})
	assert.Len(t, excluded, 1)
	assert.Len(t, included, 0)
}

func TestSendPreparedHeartbeatsRateLimitedFiltersExcludedAI(t *testing.T) {
	queueFile := filepath.Join(t.TempDir(), "offline.bdb")
	p := internalCommandParams()
	p.Offline.LastSentAt = time.Now()
	p.Offline.RateLimit = time.Hour
	p.AI.ExcludeProjects = []regex.Regex{regex.MustCompile("(?i)^nautilus$")} // nolint

	tmpDir := t.TempDir()
	aiFile := filepath.Join(tmpDir, "ai_main.py")
	require.NoError(t, os.WriteFile(aiFile, []byte("print('ai')\n"), 0644))
	codeFile := filepath.Join(tmpDir, "code_main.py")
	require.NoError(t, os.WriteFile(codeFile, []byte("print('code')\n"), 0644))

	aiHb := heartbeatpkg.Heartbeat{
		Entity:          aiFile,
		EntityType:      heartbeatpkg.FileType,
		Category:        heartbeatpkg.AICodingCategory.String(),
		ProjectOverride: "nautilus",
		Time:            1750000000,
		UserAgent:       "plugin/0.0.1",
	}
	codeHb := heartbeatpkg.Heartbeat{
		Entity:          codeFile,
		EntityType:      heartbeatpkg.FileType,
		Category:        heartbeatpkg.CodingCategory.String(),
		ProjectOverride: "nautilus",
		Time:            1750000001,
		UserAgent:       "plugin/0.0.1",
	}

	err := sendPreparedHeartbeats(t.Context(), viper.New(), p, queueFile,
		[]heartbeatpkg.Heartbeat{aiHb, codeHb}, false, loadParams)
	require.NoError(t, err)

	count, err := offlinepkg.CountHeartbeats(t.Context(), queueFile)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
