package windows_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/windows"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatFilePath(t *testing.T) {
	tests := map[string]struct {
		FilePath string
		Expected string
	}{
		"lowercase windows drive filepath": {
			FilePath: `c:\Projects\apilibrary.sl`,
			Expected: `C:/Projects/apilibrary.sl`,
		},
		"windows drive filepath with double slash": {
			FilePath: `C:\\Projects\apilibrary.sl`,
			Expected: `C:/Projects/apilibrary.sl`,
		},
		"windows remote filepath": {
			FilePath: `\\Projects\apilibrary.sl`,
			Expected: `//Projects/apilibrary.sl`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			fp, err := windows.FormatFilePath(test.FilePath)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, fp)
		})
	}
}
