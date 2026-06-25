package heartbeat_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/windows"

	"github.com/gandarez/go-realpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithFormatting(t *testing.T) {
	opt := heartbeat.WithFormatting()

	handle := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		entity, err := filepath.Abs(hh[0].Entity)
		require.NoError(t, err)

		entity, err = realpath.Realpath(entity)
		require.NoError(t, err)

		if runtime.GOOS == "windows" {
			entity = windows.FormatFilePath(entity)
		}

		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Entity: entity,
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := handle(t.Context(), []heartbeat.Heartbeat{{
		Entity:     "testdata/main.go",
		EntityType: heartbeat.FileType,
	}})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestWithFormatting_SkipsNonFileAndRemote(t *testing.T) {
	opt := heartbeat.WithFormatting()

	input := []heartbeat.Heartbeat{
		{
			Entity:     "relative-app",
			EntityType: heartbeat.AppType,
		},
		{
			Entity:     "ssh://example.com/project/main.go",
			EntityType: heartbeat.FileType,
		},
	}

	handle := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, input, hh)

		return []heartbeat.Result{{Status: 201}}, nil
	})

	result, err := handle(t.Context(), input)
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{{Status: 201}}, result)
}

func TestFormat_NotFileType(t *testing.T) {
	tests := map[string]heartbeat.EntityType{
		"app":    heartbeat.AppType,
		"domain": heartbeat.DomainType,
		"event":  heartbeat.EventType,
		"url":    heartbeat.URLType,
	}

	for name, entityType := range tests {
		t.Run(name, func(t *testing.T) {
			h := heartbeat.Heartbeat{
				Entity:     "/unmodified",
				EntityType: entityType,
			}

			formatted := heartbeat.Format(t.Context(), h)

			assert.Equal(t, "/unmodified", formatted.Entity)
		})
	}
}

func TestFormat_ProjectPathOverride(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("project path override realpath behavior is covered by Windows-specific formatting tests")
	}

	tmpDir := t.TempDir()
	entity := filepath.Join(tmpDir, "main.go")
	projectDir := filepath.Join(tmpDir, "project")
	require.NoError(t, os.Mkdir(projectDir, 0700))
	require.NoError(t, os.WriteFile(entity, []byte("package main\n"), 0600))

	h := heartbeat.Heartbeat{
		Entity:              entity,
		EntityType:          heartbeat.FileType,
		ProjectPathOverride: projectDir,
	}

	formatted := heartbeat.Format(t.Context(), h)

	entityRealpath, err := realpath.Realpath(entity)
	require.NoError(t, err)
	projectRealpath, err := realpath.Realpath(projectDir)
	require.NoError(t, err)

	assert.Equal(t, entityRealpath, formatted.Entity)
	assert.Equal(t, projectRealpath, formatted.ProjectPathOverride)
}

func TestFormat_MissingRealpathTargets(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("realpath failure branches are not used by Windows formatting")
	}

	tmpDir := t.TempDir()
	missingEntity := filepath.Join(tmpDir, "missing.go")
	formatted := heartbeat.Format(t.Context(), heartbeat.Heartbeat{
		Entity:     missingEntity,
		EntityType: heartbeat.FileType,
	})

	entityAbs, err := filepath.Abs(missingEntity)
	require.NoError(t, err)
	assert.Equal(t, entityAbs, formatted.Entity)

	entity := filepath.Join(tmpDir, "main.go")
	require.NoError(t, os.WriteFile(entity, []byte("package main\n"), 0600))

	missingProject := filepath.Join(tmpDir, "missing-project")
	formatted = heartbeat.Format(t.Context(), heartbeat.Heartbeat{
		Entity:              entity,
		EntityType:          heartbeat.FileType,
		ProjectPathOverride: missingProject,
	})

	projectAbs, err := filepath.Abs(missingProject)
	require.NoError(t, err)
	assert.Equal(t, projectAbs, formatted.ProjectPathOverride)
}

func TestFormat_WindowsPath(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping because OS is not windows.")
	}

	h := heartbeat.Heartbeat{
		Entity:     `C:\Users\project\main.go`,
		EntityType: heartbeat.FileType,
	}

	formatted := heartbeat.Format(t.Context(), h)

	assert.Equal(t, "C:/Users/project/main.go", formatted.Entity)
}

func TestFormat_NetworkMount(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping because OS is not windows.")
	}

	h := heartbeat.Heartbeat{
		Entity:     `\\192.168.1.1\apilibrary.sl`,
		EntityType: heartbeat.FileType,
	}

	r := heartbeat.Format(t.Context(), h)

	assert.Equal(t, heartbeat.Heartbeat{
		Entity:     `\\192.168.1.1/apilibrary.sl`,
		EntityType: heartbeat.FileType,
	}, r)
}
