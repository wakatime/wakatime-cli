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
				Override:    test.Override,
				Alternative: test.Alternative,
			})

			handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
				assert.Equal(t, []heartbeat.Heartbeat{
					test.Expected,
				}, hh)

				return []heartbeat.Result{
					{
						Status: 201,
					},
				}, nil
			})

			result, err := handle(test.Heartbeats)
			require.NoError(t, err)

			assert.Equal(t, []heartbeat.Result{
				{
					Status: 201,
				},
			}, result)
		})
	}
}

func TestDetect_FileDetected(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	entity := path.Join(wd, "testdata/entity.any")

	project, branch := project.Detect(entity, []project.Pattern{})

	assert.Equal(t, "wakatime-cli", project)
	assert.Equal(t, "master", branch)
}

func TestDetect_MapDetected(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	tmpFile, err := ioutil.TempFile(tmpDir, "waka-billing")
	require.NoError(t, err)

	patterns := []project.Pattern{
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

	project, branch := project.Detect(tmpFile.Name(), patterns)

	assert.Equal(t, "my-billing-project", project)
	assert.Equal(t, "", branch)
}

func TestDetect_NoneDetected(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	project, branch := project.Detect(tmpFile.Name(), []project.Pattern{})

	assert.Equal(t, "", project)
	assert.Equal(t, "", branch)
}

func testHeartbeat() heartbeat.Heartbeat {
	return heartbeat.Heartbeat{
		EntityType: heartbeat.AppType,
	}
}
