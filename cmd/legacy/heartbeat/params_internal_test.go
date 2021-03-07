package heartbeat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEditorFromPlugin(t *testing.T) {
	tests := map[string]struct {
		Plugin   string
		Expected string
	}{
		"editor/version plugin/version": {
			Plugin:   "vscode/1.51.1 vscode-wakatime/4.0.9",
			Expected: "vscode",
		},
		"plugin/version (no dash)": {
			Plugin:   "emacs-wakatime/1.0.2",
			Expected: "emacs",
		},
		"plugin/version (multiple dashes)": {
			Plugin:   "camunda-modeler-wakatime/0.4.3",
			Expected: "camunda-modeler",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			editor, err := parseEditorFromPlugin(test.Plugin)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, editor)
		})
	}
}

func TestParseEditorFromPluginErr(t *testing.T) {
	editor, err := parseEditorFromPlugin("editor-wakatime")
	require.Error(t, err)

	assert.Empty(t, editor)
	assert.Equal(t, "plugin malformed: editor-wakatime", err.Error())
}
