package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
