package heartbeat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHideProjectFolder(t *testing.T) {
	tests := map[string]struct {
		ProjectPath         string
		ProjectPathOverride string
		Entity              string
		Expected            string
	}{
		"auto-detected": {
			ProjectPath: "/usr/temp",
			Entity:      "/usr/temp/project/main.go",
			Expected:    "project/main.go",
		},
		"override": {
			ProjectPath:         "/original/folder",
			ProjectPathOverride: "/usr/temp",
			Entity:              "/usr/temp/project/main.go",
			Expected:            "project/main.go",
		},
		"windows path": {
			ProjectPath: `C:/Users/wakatime/programming/study/learn_flutter`,
			Entity:      `C:/Users/wakatime/programming/study/learn_flutter/lib/main.dart`,
			Expected:    `lib/main.dart`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			h := hideProjectFolder(Heartbeat{
				Entity:              test.Entity,
				ProjectPath:         test.ProjectPath,
				ProjectPathOverride: test.ProjectPathOverride,
			}, true)

			assert.Equal(t, test.Expected, h.Entity)
		})
	}
}
