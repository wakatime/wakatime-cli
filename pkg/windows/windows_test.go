package windows_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/windows"

	"github.com/stretchr/testify/assert"
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
			Expected: `\\Projects/apilibrary.sl`,
		},
		"windows remote ip address v4": {
			FilePath: `\\192.168.1.1\apilibrary.sl`,
			Expected: `\\192.168.1.1/apilibrary.sl`,
		},
		"windows remote ip address v6": {
			FilePath: `\\fe80::cdaf:f1ac:9c4d:6303%7\apilibrary.sl`,
			Expected: `\\fe80::cdaf:f1ac:9c4d:6303%7/apilibrary.sl`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			fp := windows.FormatFilePath(test.FilePath)

			assert.Equal(t, test.Expected, fp)
		})
	}
}
