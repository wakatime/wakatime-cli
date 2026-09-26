package deps_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/file"

	"github.com/wakatime/wakatime-cli/pkg/deps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserHTML_Parse(t *testing.T) {
	ctx := t.Context()

	tests := map[string]struct {
		Filepath string
		Expected []string
	}{
		"html": {
			Filepath: "testdata/html.html",
			Expected: []string{
				`wakatime.js`,
				`../scripts/wakatime.js`,
				`https://www.wakatime.com/scripts/my.js`,
				"this is a\n multiline value",
			},
		},
		"html django": {
			Filepath: "testdata/html_django.html",
			Expected: []string{`libs/json2.js`},
		},
		"html with PHP": {
			Filepath: "testdata/html_with_php.html",
			Expected: []string{`https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/js/bootstrap.min.js`},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parser := deps.ParserHTML{}

			dependencies, err := parser.Parse(ctx, test.Filepath)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, dependencies)
		})
	}
}

func TestParserHTML_ParseInlineContent(t *testing.T) {
	tests := map[string]struct {
		Content  string
		Expected []string
	}{
		"truncated script": {
			Content: `<script src="before.js"></script><script src="large.js">const s="` +
				strings.Repeat("a", 6*1024*1024) +
				`";</script><script src="beyond-limit.js"></script>`,
			Expected: []string{"before.js", "large.js"},
		},
		"complete large script": {
			Content: `<script>const s="` +
				strings.Repeat("a", file.MaxFileSizeSupported/2) +
				`";</script><script src="after.js"></script>`,
			Expected: []string{"after.js"},
		},
		"unterminated style": {
			Content: `<script src="before.js"></script><style>` +
				strings.Repeat("a", 6*1024*1024),
			Expected: []string{"before.js"},
		},
		"markup in raw text and comments": {
			Content: `<!-- <script src="comment.js"></script> -->` +
				`<script>const s='<script src="inline.js">';</script>` +
				`<style>/* <script src="style.js"> */</style>` +
				`<textarea><script src="textarea.js"></script></textarea>` +
				`<script src="real.js"></script>`,
			Expected: []string{"real.js"},
		},
		"attributes": {
			Content: `<SCRIPT SRC='single.js'></SCRIPT>` +
				`<script src=unquoted.js></script>` +
				`<script data-src="ignored.js" src="a.js?x=1&amp;y=2"></script>`,
			Expected: []string{"single.js", "unquoted.js", "a.js?x=1&y=2"},
		},
		"incomplete tag": {
			Content:  `<script src="before.js"></script><script src="unfinished`,
			Expected: []string{"before.js"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "fixture.html")
			require.NoError(t, os.WriteFile(path, []byte(test.Content), 0600))

			parser := deps.ParserHTML{}
			actual, err := parser.Parse(t.Context(), path)
			require.NoError(t, err)
			assert.Equal(t, test.Expected, actual)
		})
	}
}
