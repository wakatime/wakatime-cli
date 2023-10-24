package lexer_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
)

func TestPovRay_AnalyseText(t *testing.T) {
	tests := map[string]struct {
		Filepath string
		Expected float32
	}{
		"camera": {
			Filepath: "testdata/povray_camera.pov",
			Expected: 0.05,
		},
		"light_source": {
			Filepath: "testdata/povray_light_source.pov",
			Expected: 0.1,
		},
		"declare": {
			Filepath: "testdata/povray_declare.pov",
			Expected: 0.05,
		},
		"version": {
			Filepath: "testdata/povray_version.pov",
			Expected: 0.05,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(test.Filepath)
			assert.NoError(t, err)

			l := lexer.POVRay{}.Lexer()

			assert.Equal(t, test.Expected, l.AnalyseText(string(data)))
		})
	}
}
