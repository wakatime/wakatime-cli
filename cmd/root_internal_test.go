package cmd

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/params"
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

func TestGrokBuildAsEditorFlag(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		setting bool
		want    bool
	}{
		{name: "default"},
		{name: "config enabled", setting: true, want: true},
		{name: "flag enabled", args: []string{"--grok-build-as-editor"}, want: true},
		{name: "flag disabled overrides config", args: []string{"--grok-build-as-editor=false"}, setting: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := vipertools.MustNew()
			v.Set("settings.grok_build_as_editor", tt.setting)

			command := &cobra.Command{}
			setFlags(command, v)
			require.NoError(t, command.ParseFlags(tt.args))

			got, err := params.LoadAIParams(t.Context(), v, params.FlagReadOrderFlagPrecedence)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.GrokBuildAsEditor)
		})
	}
}
