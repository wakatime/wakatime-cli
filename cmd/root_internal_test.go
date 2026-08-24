package cmd

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncAIActivityEnabled(t *testing.T) {
	tests := map[string]string{
		"activity flag":    "--sync-ai-activity",
		"heartbeats alias": "--sync-ai-heartbeats",
	}

	for name, flag := range tests {
		t.Run(name, func(t *testing.T) {
			v := vipertools.MustNew()
			command := &cobra.Command{}
			setFlags(command, v)

			require.NoError(t, command.ParseFlags([]string{flag}))
			assert.True(t, syncAIActivityEnabled(v))
		})
	}
}

func TestSyncAIHeartbeatsFlagHidden(t *testing.T) {
	v := vipertools.MustNew()
	command := &cobra.Command{}
	setFlags(command, v)

	flag := command.Flags().Lookup("sync-ai-heartbeats")
	require.NotNil(t, flag)
	assert.True(t, flag.Hidden)
}

func TestCodexSourceFlag(t *testing.T) {
	v := vipertools.MustNew()
	command := &cobra.Command{}
	setFlags(command, v)

	flag := command.Flags().Lookup("codex-source")
	require.NotNil(t, flag)
	assert.True(t, flag.Hidden)

	require.NoError(t, command.ParseFlags([]string{"--codex-source", "kandev"}))
	assert.Equal(t, "kandev", v.GetString("codex-source"))
}
