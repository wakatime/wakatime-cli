package project

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/gandarez/go-realpath"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/log/setup"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"
)

func TestObfuscateProjectName(t *testing.T) {
	tmpDir, err := realpath.Realpath(t.TempDir())
	require.NoError(t, err)

	project := obfuscateProjectName(t.Context(), tmpDir)

	assert.NotEmpty(t, project)
}

func TestObfuscateProjectName_EmptyFolder(t *testing.T) {
	project := obfuscateProjectName(t.Context(), "")

	assert.Empty(t, project)
}

func TestObfuscateProjectName_WakatimeProjectTakesPrecedence(t *testing.T) {
	tmpDir, err := realpath.Realpath(t.TempDir())
	require.NoError(t, err)

	copyFile(
		t,
		"testdata/wakatime-project",
		filepath.Join(tmpDir, ".wakatime-project"),
	)

	project := obfuscateProjectName(t.Context(), tmpDir)

	assert.Empty(t, project)
}

func TestObfuscateProjectName_Write_Err(t *testing.T) {
	ctx := t.Context()

	logFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer logFile.Close()

	v := vipertools.MustNew()
	v.Set("log-file", logFile.Name())
	v.Set("verbose", true)

	logger, err := setup.Logging(ctx, v)
	require.NoError(t, err)

	defer logger.Flush()

	ctx = log.ToContext(ctx, logger)

	project := obfuscateProjectName(ctx, "-invalid-folder-name-")

	assert.Empty(t, project)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "failed to write: failed to save wakatime project file")
}

func copyFile(t *testing.T, source, destination string) {
	input, err := os.ReadFile(source)
	require.NoError(t, err)

	err = os.WriteFile(destination, input, 0600)
	require.NoError(t, err)
}

func TestInterpolateProjectPlaceholder_WithVCSProject(t *testing.T) {
	result := interpolateProjectPlaceholder("my-company/{project}", "wakatime-cli", "/some/folder")

	assert.Equal(t, "my-company/wakatime-cli", result)
}

func TestInterpolateProjectPlaceholder_WithVCSProject_AsPrefix(t *testing.T) {
	result := interpolateProjectPlaceholder("{project}-internal", "wakatime-cli", "/some/folder")

	assert.Equal(t, "wakatime-cli-internal", result)
}

func TestInterpolateProjectPlaceholder_WithVCSProject_Alone(t *testing.T) {
	result := interpolateProjectPlaceholder("{project}", "wakatime-cli", "/some/folder")

	assert.Equal(t, "wakatime-cli", result)
}

func TestInterpolateProjectPlaceholder_WithVCSProject_MultiplePlaceholders(t *testing.T) {
	result := interpolateProjectPlaceholder("{project}/{project}", "wakatime-cli", "/some/folder")

	assert.Equal(t, "wakatime-cli/wakatime-cli", result)
}

func TestInterpolateProjectPlaceholder_FallbackToFolderBasename(t *testing.T) {
	result := interpolateProjectPlaceholder("my-company/{project}", "", "/path/to/my-project")

	assert.Equal(t, "my-company/my-project", result)
}

func TestInterpolateProjectPlaceholder_FallbackToFolderBasename_Alone(t *testing.T) {
	result := interpolateProjectPlaceholder("{project}", "", "/path/to/my-project")

	assert.Equal(t, "my-project", result)
}

func TestInterpolateProjectPlaceholder_EmptyVCSAndFolder(t *testing.T) {
	result := interpolateProjectPlaceholder("my-company/{project}", "", "")

	assert.Equal(t, "my-company/{project}", result)
}

func TestInterpolateProjectPlaceholder_FolderIsDot(t *testing.T) {
	result := interpolateProjectPlaceholder("my-company/{project}", "", ".")

	assert.Equal(t, "my-company/{project}", result)
}

func TestInterpolateProjectPlaceholder_FolderIsSlash(t *testing.T) {
	result := interpolateProjectPlaceholder("my-company/{project}", "", "/")

	assert.Equal(t, "my-company/{project}", result)
}

func TestInterpolateProjectPlaceholder_NoPlaceholder(t *testing.T) {
	result := interpolateProjectPlaceholder("my-static-project", "wakatime-cli", "/some/folder")

	assert.Equal(t, "my-static-project", result)
}

func TestInterpolateProjectPlaceholder_VCSProjectTakesPrecedenceOverFolder(t *testing.T) {
	// Even if folder has a different name, VCS project should be used
	result := interpolateProjectPlaceholder("prefix-{project}", "vcs-name", "/path/to/folder-name")

	assert.Equal(t, "prefix-vcs-name", result)
}

func TestProjectHelperBranches(t *testing.T) {
	tmpDir := t.TempDir()
	childDir := filepath.Join(tmpDir, "child", "grandchild")
	require.NoError(t, os.MkdirAll(childDir, 0700))

	marker := filepath.Join(tmpDir, ".marker")
	require.NoError(t, os.WriteFile(marker, []byte("ok"), 0600))

	found, ok := FindFileOrDirectory(t.Context(), childDir, ".marker")
	require.True(t, ok)
	assert.Equal(t, marker, found)

	found, ok = FindFileOrDirectory(t.Context(), tmpDir, ".missing")
	assert.False(t, ok)
	assert.Empty(t, found)

	assert.True(t, isRootPath(""))
	assert.True(t, isRootPath("."))
	assert.Equal(t, "", firstNonEmptyString("", ""))
	assert.Equal(t, "second", firstNonEmptyString("", "second", "third"))
	assert.Empty(t, FormatProjectFolder(t.Context(), ""))
	assert.Equal(
		t,
		"template/{project}",
		interpolateProjectPlaceholder("template/{project}", "", string(filepath.Separator)),
	)
	assert.Positive(t, CountSlashesInProjectFolder(filepath.Join(tmpDir, "child")))
	assert.Equal(
		t,
		"repo",
		resolveSvnInfo(map[string]string{
			"Repository Root": "https://svn.example.com/team/repo\r",
		}, "Repository Root"),
	)
	assert.Equal(t, "branch", resolveSvnInfo(map[string]string{"URL": `https://svn.example.com\team\branch`}, "URL"))
	assert.Empty(t, resolveSvnInfo(map[string]string{}, "URL"))
}
