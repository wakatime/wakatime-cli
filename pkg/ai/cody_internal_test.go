//go:build !freebsd && !openbsd && !netbsd && !dragonfly

package ai

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodyContextItemFilePathHandlesWindowsFilePaths(t *testing.T) {
	for _, tc := range []struct {
		name string
		item codyContextItem
		want string
	}{
		{
			name: "fs path",
			item: codyContextItem{URI: codyURI{FSPath: `C:\Users\runner\project\main.go`}},
			want: `C:\Users\runner\project\main.go`,
		},
		{
			name: "file uri with drive",
			item: codyContextItem{URI: codyURI{Path: `file://C:\Users\runner\project\main.go`}},
			want: `C:\Users\runner\project\main.go`,
		},
		{
			name: "vscode file uri path",
			item: codyContextItem{URI: codyURI{Scheme: "file", Path: `/C:/Users/runner/project/main.go`}},
			want: filepath.FromSlash(`C:/Users/runner/project/main.go`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.item.filePath())
		})
	}
}
