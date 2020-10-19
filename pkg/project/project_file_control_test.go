package project_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "wakatime-git")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	fc := project.FileControl{
		Path:    tmpDir,
		Project: "billing",
	}

	err = fc.Write()
	require.NoError(t, err)

	expected, err := ioutil.ReadFile("testdata/.wakatime-project-only-project")
	require.NoError(t, err)

	actual, err := ioutil.ReadFile(path.Join(fc.Path, ".wakatime-project"))
	require.NoError(t, err)

	assert.Equal(t, string(expected), string(actual))
}
