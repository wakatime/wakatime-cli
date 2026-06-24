package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/version"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCMD(t *testing.T) {
	command := NewRootCMD()

	assert.Equal(t, "wakatime-cli", command.Use)
	assert.Equal(t, "Command line interface used by all WakaTime text editor plugins.", command.Short)
	require.NotNil(t, command.RunE)

	require.NotNil(t, command.Flags().Lookup("entity"))
	require.NotNil(t, command.Flags().Lookup("sync-ai-heartbeats"))
	assert.True(t, command.Flags().Lookup("sync-ai-heartbeats").Hidden)
}

func TestRunE_DispatchesLocalCommands(t *testing.T) {
	t.Run("user agent", func(t *testing.T) {
		v := newRunEViper(t)
		v.Set("plugin", "vim/1.0")
		v.Set("user-agent", true)

		var err error

		output := captureStdout(t, func() {
			err = RunE(newRunECommand(), v)
		})

		require.NoError(t, err)
		assert.Contains(t, output, "vim/1.0")
	})

	t.Run("version", func(t *testing.T) {
		oldVersion := version.Version
		version.Version = "1.2.3"

		t.Cleanup(func() { version.Version = oldVersion })

		v := newRunEViper(t)
		v.Set("version", true)

		var err error

		output := captureStdout(t, func() {
			err = RunE(newRunECommand(), v)
		})

		require.NoError(t, err)
		assert.Equal(t, "1.2.3\n", output)
	})

	t.Run("config read", func(t *testing.T) {
		v := newRunEViper(t)
		v.Set("config-read", "api_key")
		v.Set("config-section", "settings")
		v.Set("settings.api_key", "secret")

		var err error

		output := captureStdout(t, func() {
			err = RunE(newRunECommand(), v)
		})

		require.NoError(t, err)
		assert.Equal(t, "secret\n", output)
	})

	t.Run("config write", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), ".wakatime.cfg")
		v := newRunEViper(t)
		v.Set("config", configPath)
		v.Set("config-write", map[string]string{"debug": "true"})
		v.Set("config-section", "settings")

		var err error

		output := captureStdout(t, func() {
			err = RunE(newRunECommand(), v)
		})

		require.NoError(t, err)
		assert.Empty(t, output)

		config, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Contains(t, string(config), "debug")
		assert.Contains(t, string(config), "true")
	})
}

func TestRunE_DispatchesErrorCommands(t *testing.T) {
	tests := map[string]struct {
		Configure func(t *testing.T, v *viper.Viper)
		ExitCode  int
	}{
		"today": {
			Configure: func(_ *testing.T, v *viper.Viper) {
				v.Set("today", true)
			},
			ExitCode: exitcode.ErrGeneric,
		},
		"today goal": {
			Configure: func(_ *testing.T, v *viper.Viper) {
				v.Set("today-goal", "invalid")
			},
			ExitCode: exitcode.ErrGeneric,
		},
		"file experts": {
			Configure: func(_ *testing.T, v *viper.Viper) {
				v.Set("file-experts", true)
			},
			ExitCode: exitcode.ErrGeneric,
		},
		"heartbeat": {
			Configure: func(t *testing.T, v *viper.Viper) {
				entity := filepath.Join(t.TempDir(), "main.go")
				require.NoError(t, os.WriteFile(entity, []byte("package main\n"), 0600))

				v.Set("entity", entity)
				v.Set("offline-queue-file", filepath.Join(t.TempDir(), "offline.bdb"))
				v.Set("sync-ai-disabled", true)
			},
			ExitCode: exitcode.ErrAuth,
		},
		"sync offline activity": {
			Configure: func(t *testing.T, v *viper.Viper) {
				v.Set("offline-queue-file", filepath.Join(t.TempDir(), "offline.bdb"))
				v.Set("sync-offline-activity", 1)
			},
			ExitCode: exitcode.ErrAuth,
		},
		"sync ai activity": {
			Configure: func(_ *testing.T, v *viper.Viper) {
				v.Set("sync-ai-activity", true)
			},
			ExitCode: exitcode.ErrAuth,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := newRunEViper(t)
			test.Configure(t, v)

			err := RunE(newRunECommand(), v)

			var errexitcode exitcode.Err
			require.ErrorAs(t, err, &errexitcode)
			assert.Equal(t, test.ExitCode, errexitcode.Code)
		})
	}
}

func TestRunE_DispatchesOfflineOutputCommands(t *testing.T) {
	t.Run("offline count", func(t *testing.T) {
		v := newRunEViper(t)
		v.Set("offline-count", true)
		v.Set("offline-queue-file", filepath.Join(t.TempDir(), "offline.bdb"))

		var err error

		output := captureStdout(t, func() {
			err = RunE(newRunECommand(), v)
		})

		require.NoError(t, err)
		assert.Equal(t, "0\n", output)
	})

	t.Run("print offline heartbeats", func(t *testing.T) {
		v := newRunEViper(t)
		v.Set("print-offline-heartbeats", 1)
		v.Set("offline-queue-file", filepath.Join(t.TempDir(), "offline.bdb"))

		var err error

		output := captureStdout(t, func() {
			err = RunE(newRunECommand(), v)
		})

		require.NoError(t, err)
		assert.Equal(t, "[]\n", output)
	})
}

func TestRunE_NoCommand(t *testing.T) {
	err := RunE(newRunECommand(), newRunEViper(t))

	var errexitcode exitcode.Err
	require.ErrorAs(t, err, &errexitcode)
	assert.Equal(t, exitcode.ErrGeneric, errexitcode.Code)
}

func TestRunE_ConfigParseErrorFallsBackToOfflineQueue(t *testing.T) {
	tmpDir := t.TempDir()
	entity := filepath.Join(tmpDir, "main.go")
	require.NoError(t, os.WriteFile(entity, []byte("package main\n"), 0600))

	queueFilepath := filepath.Join(tmpDir, "offline.bdb")

	v := newRunEViper(t)
	v.Set("config", tmpDir)
	v.Set("entity", entity)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("offline-queue-file", queueFilepath)
	v.Set("sync-ai-disabled", true)

	err := RunE(newRunECommand(), v)

	var errexitcode exitcode.Err
	require.ErrorAs(t, err, &errexitcode)
	assert.Equal(t, exitcode.ErrConfigFileParse, errexitcode.Code)

	count, err := offline.CountHeartbeats(t.Context(), queueFilepath)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestRunCmdWithOfflineSync_Success(t *testing.T) {
	v := vipertools.MustNew()
	v.Set("disable-offline", true)

	err := RunCmdWithOfflineSync(t.Context(), v, false, false, func(_ context.Context, _ *viper.Viper) (int, error) {
		return exitcode.Success, nil
	})

	require.NoError(t, err)
}

func TestRunVersionVerbose(t *testing.T) {
	oldVersion := version.Version
	oldCommit := version.Commit
	oldBuildDate := version.BuildDate
	oldOS := version.OS
	oldArch := version.Arch

	t.Cleanup(func() {
		version.Version = oldVersion
		version.Commit = oldCommit
		version.BuildDate = oldBuildDate
		version.OS = oldOS
		version.Arch = oldArch
	})

	version.Version = "1.2.3"
	version.Commit = "abcdef"
	version.BuildDate = "2026-06-24"
	version.OS = "testos"
	version.Arch = "testarch"

	v := vipertools.MustNew()
	v.Set("verbose", true)

	var (
		code int
		err  error
	)

	output := captureStdout(t, func() {
		code, err = runVersion(t.Context(), v)
	})

	require.NoError(t, err)
	assert.Equal(t, exitcode.Success, code)
	assert.Equal(t,
		"wakatime-cli\n"+
			"  Version: 1.2.3\n"+
			"  Commit: abcdef\n"+
			"  Built: 2026-06-24\n"+
			"  OS/Arch: testos/testarch\n",
		output,
	)
}

func newRunEViper(t *testing.T) *viper.Viper {
	t.Helper()

	tmpDir := t.TempDir()
	v := vipertools.MustNew()
	v.Set("config", filepath.Join(tmpDir, "missing.cfg"))
	v.Set("internal-config", filepath.Join(tmpDir, "missing-internal.cfg"))
	v.Set("log-file", filepath.Join(tmpDir, "wakatime.log"))

	return v
}

func newRunECommand() *cobra.Command {
	command := &cobra.Command{Use: "wakatime-cli"}
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)

	return command
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	stdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = w

	defer func() { os.Stdout = stdout }()

	outC := make(chan string, 1)
	errC := make(chan error, 1)

	go func() {
		var buf bytes.Buffer

		_, err := io.Copy(&buf, r)
		errC <- err

		outC <- buf.String()
	}()

	fn()

	require.NoError(t, w.Close())

	output := <-outC

	require.NoError(t, <-errC)
	require.NoError(t, r.Close())

	return output
}
