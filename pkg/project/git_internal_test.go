package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHelperBranches(t *testing.T) {
	ctx := t.Context()

	_, err := findGitdir(ctx, filepath.Join(t.TempDir(), "missing-git-file"))
	require.Error(t, err)

	root := t.TempDir()
	gitdir := filepath.Join(root, ".git")
	require.NoError(t, os.MkdirAll(gitdir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(gitdir, "HEAD"), []byte("ref: refs/heads/main\n"), 0600))

	resolved, err := resolveGitdir(root, ".git")
	require.NoError(t, err)
	assert.Equal(t, gitdir, resolved)

	resolved, err = resolveGitdir(root, filepath.Join(root, "missing"))
	require.NoError(t, err)
	assert.Empty(t, resolved)

	commondir, ok, err := findCommondir(ctx, "")
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, commondir)

	commondir, ok, err = findCommondir(ctx, gitdir)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, commondir)

	worktreeGitdir := filepath.Join(root, ".git", "worktrees", "feature")
	require.NoError(t, os.MkdirAll(worktreeGitdir, 0700))
	commondir, ok, err = findCommondir(ctx, worktreeGitdir)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, commondir)

	require.NoError(t, os.WriteFile(filepath.Join(worktreeGitdir, "commondir"), nil, 0600))
	commondir, ok, err = resolveCommondir(ctx, worktreeGitdir)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, commondir)
}

func TestGitRemoteHelperBranches(t *testing.T) {
	ctx := t.Context()
	root := t.TempDir()

	assert.Equal(t, "project", projectOrRemote(ctx, "project", false, filepath.Join(root, ".git")))
	assert.Equal(t, "project", projectOrRemote(ctx, "project", true, filepath.Join(root, ".git")))

	configPath := filepath.Join(root, ".git", "config")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0700))
	require.NoError(t, os.WriteFile(configPath, []byte(strings.Join([]string{
		"[remote \"origin\"]",
		"	url = git@github.com:wakatime/wakatime-cli.git",
		"[branch \"main\"]",
	}, "\n")), 0600))

	remote, err := findGitRemote(ctx, configPath)
	require.NoError(t, err)
	assert.Equal(t, "wakatime/wakatime-cli", remote)
	assert.Equal(t, "wakatime/wakatime-cli", projectOrRemote(ctx, "project", true, filepath.Dir(configPath)))

	require.NoError(t, os.WriteFile(configPath, []byte(strings.Join([]string{
		"[remote \"origin\"]",
		"	url without equals",
	}, "\n")), 0600))
	remote, err = findGitRemote(ctx, configPath)
	require.NoError(t, err)
	assert.Empty(t, remote)

	require.NoError(t, os.WriteFile(configPath, []byte(strings.Join([]string{
		"[remote \"origin\"]",
		"	url = no-colon",
	}, "\n")), 0600))
	_, err = findGitRemote(ctx, configPath)
	require.Error(t, err)
}
