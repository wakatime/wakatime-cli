package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/log"
)

const defaultCountLinesChangedTimeoutSecs = 2

var gitLinesChangedRegex = regexp.MustCompile(`^(?P<added>\d+)\s*(?P<removed>\d+)\s*(?s).*$`)

type (
	// Git is an interface to git.
	Git interface {
		CountLinesChanged() (*int, *int, error)
	}

	// Client is a git client.
	Client struct {
		filepath string
		GitCmd   func(args ...string) (string, error)
	}
)

// New creates a new git client.
func New(filepath string) *Client {
	return &Client{
		filepath: filepath,
		GitCmd:   gitCmdFn,
	}
}

// gitCmdFn runs a git command with the specified env vars and returns its output or errors.
func gitCmdFn(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultCountLinesChangedTimeoutSecs*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to execute git command: %s", stderr.String())
	}

	return stdout.String(), nil
}

// CountLinesChanged counts the number of lines added and removed in a file.
func (c *Client) CountLinesChanged() (*int, *int, error) {
	if !fileExists(c.filepath) {
		return nil, nil, nil
	}

	out, err := c.GitCmd("-C", filepath.Dir(c.filepath), "diff", "--numstat", c.filepath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count lines changed: %s", err)
	}

	if out == "" {
		// Maybe it's staged, try with --cached.
		out, err = c.GitCmd("-C", filepath.Dir(c.filepath), "diff", "--numstat", "--cached", c.filepath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to count lines changed: %s", err)
		}
	}

	if out == "" {
		return nil, nil, nil
	}

	match := gitLinesChangedRegex.FindStringSubmatch(out)
	paramsMap := make(map[string]string)

	for i, name := range gitLinesChangedRegex.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}

	if len(paramsMap) == 0 {
		log.Debugf("failed to parse git diff output: %s", out)

		return nil, nil, nil
	}

	var added, removed *int

	if val, ok := paramsMap["added"]; ok {
		if v, err := strconv.Atoi(val); err == nil {
			added = &v
		}
	}

	if val, ok := paramsMap["removed"]; ok {
		if v, err := strconv.Atoi(val); err == nil {
			removed = &v
		}
	}

	return added, removed, nil
}

// fileExists checks if a file or directory exist.
func fileExists(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}
