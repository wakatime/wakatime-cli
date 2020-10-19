package project_test

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
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

func TestDetectWithDetection_OverrideTakesPrecedence(t *testing.T) {
	fp, tearDown := setupTestGitBasic(t)
	defer tearDown()

	opt := project.WithDetection(project.Config{
		Override: "billing",
	})

	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Entity:     path.Join(fp, "wakatime-cli/src/pkg/file.go"),
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
			Entity:     path.Join(fp, "wakatime-cli/src/pkg/file.go"),
		},
	})
	require.NoError(t, err)
}

func TestDetectWithDetection_ObfuscateProject(t *testing.T) {
	fp, tearDown := setupTestGitBasic(t)
	defer tearDown()

	opt := project.WithDetection(project.Config{
		ShouldObfuscateProject: true,
	})

	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.NotEmpty(t, hh[0].Project)
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Entity:     path.Join(fp, "wakatime-cli/src/pkg/file.go"),
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
			Entity:     path.Join(fp, "wakatime-cli/src/pkg/file.go"),
		},
	})
	require.NoError(t, err)
}

func TestDetect_FileDetected(t *testing.T) {
	result := project.Detect("testdata/entity.any", []project.MapPattern{})

	assert.Equal(t, project.Result{
		Project: "wakatime-cli",
		Branch:  "master",
	}, result)
}

func TestDetect_MapDetected(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	tmpFile, err := ioutil.TempFile(tmpDir, "waka-billing")
	require.NoError(t, err)

	patterns := []project.MapPattern{
		{
			Name: "my-project-1",
			Regex: func() *regexp.Regexp {
				r, err := regexp.Compile(filepath.Join(tmpDir, "path/to/otherfolder"))
				require.NoError(t, err)
				return r
			}(),
		},
		{
			Name: "my-{0}-project",
			Regex: func() *regexp.Regexp {
				r, err := regexp.Compile(filepath.Join(tmpDir, "waka-([a-z]+)"))
				require.NoError(t, err)
				return r
			}(),
		},
	}

	result := project.Detect(tmpFile.Name(), patterns)

	assert.Equal(t, project.Result{
		Project: "my-billing-project",
	}, result)
}

func TestDetectWithRevControl_GitDetected(t *testing.T) {
	fp, tearDown := setupTestGitBasic(t)
	defer tearDown()

	result := project.DetectWithRevControl(
		path.Join(fp, "wakatime-cli/src/pkg/file.go"),
		[]*regexp.Regexp{}, false, "", "")

	assert.Contains(t, result.Folder, path.Join(fp, "wakatime-cli"))
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

	result := project.Detect(tmpFile.Name(), []project.MapPattern{})

	assert.Equal(t, project.Result{}, result)
}

func testHeartbeat() heartbeat.Heartbeat {
	return heartbeat.Heartbeat{
		EntityType: heartbeat.AppType,
	}
}
