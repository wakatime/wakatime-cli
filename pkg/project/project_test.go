package project_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithDetection_EntityNotFile(t *testing.T) {
	tests := map[string]struct {
		Heartbeats  []heartbeat.Heartbeat
		Override    string
		Alternative string
		Expected    heartbeat.Heartbeat
	}{
		"entity not file override takes precedence": {
			Heartbeats:  []heartbeat.Heartbeat{testHeartbeat()},
			Override:    "billing",
			Alternative: "pci",
			Expected: heartbeat.Heartbeat{
				EntityType: heartbeat.AppType,
				Project:    heartbeat.String("billing"),
			},
		},
		"entity not file alternative takes precedence": {
			Heartbeats:  []heartbeat.Heartbeat{testHeartbeat()},
			Alternative: "pci",
			Expected: heartbeat.Heartbeat{
				EntityType: heartbeat.AppType,
				Project:    heartbeat.String("pci"),
			},
		},
		"entity not file empty return": {
			Heartbeats: []heartbeat.Heartbeat{testHeartbeat()},
			Expected: heartbeat.Heartbeat{
				EntityType: heartbeat.AppType,
				Project:    heartbeat.String(""),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			opt := project.WithDetection(project.Config{
				Override:  test.Override,
				Alternate: test.Alternative,
			})

			handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
				assert.Equal(t, []heartbeat.Heartbeat{
					test.Expected,
				}, hh)

				return nil, nil
			})

			_, err := handle(test.Heartbeats)
			require.NoError(t, err)
		})
	}
}

func TestWithDetection_OverrideTakesPrecedence(t *testing.T) {
	fp, tearDown := setupTestGitBasic(t)
	defer tearDown()

	opt := project.WithDetection(project.Config{
		Override: "billing",
	})

	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Entity:     filepath.Join(fp, "wakatime-cli/src/pkg/file.go"),
				EntityType: heartbeat.FileType,
				Project:    heartbeat.String("billing"),
				Branch:     heartbeat.String("master"),
			},
		}, hh)

		return nil, nil
	})

	_, err := handle([]heartbeat.Heartbeat{
		{
			EntityType: heartbeat.FileType,
			Entity:     filepath.Join(fp, "wakatime-cli/src/pkg/file.go"),
		},
	})
	require.NoError(t, err)
}

func TestWithDetection_ObfuscateProject(t *testing.T) {
	fp, tearDown := setupTestGitBasic(t)
	defer tearDown()

	opt := project.WithDetection(project.Config{
		ShouldObfuscateProject: true,
	})

	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.NotEmpty(t, hh[0].Project)
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Entity:     filepath.Join(fp, "wakatime-cli/src/pkg/file.go"),
				EntityType: heartbeat.FileType,
				Project:    hh[0].Project,
				Branch:     heartbeat.String("master"),
			},
		}, hh)

		return nil, nil
	})

	_, err := handle([]heartbeat.Heartbeat{
		{
			EntityType: heartbeat.FileType,
			Entity:     filepath.Join(fp, "wakatime-cli/src/pkg/file.go"),
		},
	})
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(fp, "wakatime-cli/.wakatime-project"))
}

func TestDetect_FileDetected(t *testing.T) {
	project, branch := project.Detect("testdata/entity.any", []project.MapPattern{})

	assert.Equal(t, "wakatime-cli", project)
	assert.Equal(t, "master", branch)
}

func TestDetect_MapDetected(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	tmpFile, err := ioutil.TempFile(tmpDir, "waka-billing")
	require.NoError(t, err)

	patterns := []project.MapPattern{
		{
			Name:  "my-project-1",
			Regex: regexp.MustCompile(formatRegex(filepath.Join(tmpDir, "path", "to", "otherfolder"))),
		},
		{
			Name:  "my-{0}-project",
			Regex: regexp.MustCompile(formatRegex(filepath.Join(tmpDir, "waka-([a-z]+)"))),
		},
	}

	project, branch := project.Detect(tmpFile.Name(), patterns)

	assert.Equal(t, "my-billing-project", project)
	assert.Empty(t, branch)
}

func TestDetectWithRevControl_GitDetected(t *testing.T) {
	fp, tearDown := setupTestGitBasic(t)
	defer tearDown()

	result := project.DetectWithRevControl(
		filepath.Join(fp, "wakatime-cli/src/pkg/file.go"),
		[]*regexp.Regexp{}, false)

	assert.Contains(t, result.Folder, filepath.Join(fp, "wakatime-cli"))
	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "master",
		Folder:  result.Folder,
	}, result)
}

func TestDetect_NoProjectDetected(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	project, branch := project.Detect(tmpFile.Name(), []project.MapPattern{})

	assert.Empty(t, project)
	assert.Empty(t, branch)
}

func testHeartbeat() heartbeat.Heartbeat {
	return heartbeat.Heartbeat{
		EntityType: heartbeat.AppType,
	}
}

func formatRegex(fp string) string {
	if runtime.GOOS != "windows" {
		return fp
	}

	return strings.ReplaceAll(fp, `\`, `\\`)
}
