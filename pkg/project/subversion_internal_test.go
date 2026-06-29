package project

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveSvnInfo(t *testing.T) {
	info := map[string]string{
		"Repository Root": "file:///D:/temp/SVN/wakatime-cli\r",
		"URL":             "file:///D:/temp/SVN/wakatime-cli/branches/billing",
		"Windows URL":     `file:///D:\temp\SVN\wakatime-cli\trunk`,
	}

	assert.Equal(t, "wakatime-cli", resolveSvnInfo(info, "Repository Root"))
	assert.Equal(t, "billing", resolveSvnInfo(info, "URL"))
	assert.Equal(t, "trunk", resolveSvnInfo(info, "Windows URL"))
	assert.Empty(t, resolveSvnInfo(info, "missing"))
}

func TestGitID(t *testing.T) {
	assert.Equal(t, GitDetector, Git{}.ID())
}

func TestSubversionDetectWithInjectedCommands(t *testing.T) {
	commands := svnCommands{
		version: func(loc string) error {
			if loc == "svn" {
				return nil
			}

			return errors.New("missing")
		},
		info: func(binary, fp string) ([]byte, error) {
			assert.Equal(t, "svn", binary)
			assert.True(t, filepath.IsAbs(fp))

			return []byte(strings.Join([]string{
				"Repository Root: file:///D:/temp/SVN/wakatime-cli",
				"URL: file:///D:/temp/SVN/wakatime-cli/branches/billing",
			}, "\n")), nil
		},
		xcodeToolsExist: func() bool { return true },
	}

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".svn"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".svn", "wc.db"), nil, 0600))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "src"), 0700))
	entity := filepath.Join(root, "src", "main.go")
	require.NoError(t, os.WriteFile(entity, []byte("package main\n"), 0600))

	result, detected, err := (Subversion{Filepath: entity}).detect(t.Context(), commands)

	require.NoError(t, err)
	assert.True(t, detected)
	assert.Equal(t, Result{
		Project: "wakatime-cli",
		Branch:  "billing",
		Folder:  "file:///D:/temp/SVN/wakatime-cli",
	}, result)
}

func TestSubversionDetectInfoError(t *testing.T) {
	commands := svnCommands{
		version:         func(string) error { return nil },
		info:            func(string, string) ([]byte, error) { return nil, errors.New("svn failed") },
		xcodeToolsExist: func() bool { return true },
	}

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".svn"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".svn", "wc.db"), nil, 0600))
	entity := filepath.Join(root, "main.go")
	require.NoError(t, os.WriteFile(entity, []byte("package main\n"), 0600))

	_, detected, err := (Subversion{Filepath: entity}).detect(t.Context(), commands)

	require.Error(t, err)
	assert.False(t, detected)
	assert.Contains(t, err.Error(), "failed to get svn info")
}

func TestFindSvnBinaryInjected(t *testing.T) {
	var checked []string

	commands := svnCommands{
		version: func(loc string) error {
			checked = append(checked, loc)
			if loc == "/usr/local/bin/svn" {
				return nil
			}

			return errors.New("missing")
		},
	}

	binary, ok := findSvnBinary(commands)

	assert.True(t, ok)
	assert.Equal(t, "/usr/local/bin/svn", binary)
	assert.Equal(t, []string{"svn", "/usr/bin/svn", "/usr/local/bin/svn"}, checked)
}

func TestSvnInfoNoXcodeTools(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("xcode tools gate is darwin-only")
	}

	info, ok, err := svnInfo(t.TempDir(), "svn", svnCommands{
		info:            func(string, string) ([]byte, error) { return nil, nil },
		xcodeToolsExist: func() bool { return false },
	})

	require.NoError(t, err)
	assert.False(t, ok)
	assert.Nil(t, info)
}
