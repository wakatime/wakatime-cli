package windows

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	netUseOutputOne = `New connections will be remembered.

Status       Local     Remote                    Network

-------------------------------------------------------------------------------
OK           Z:        \\remotepc\share          Microsoft Windows Network
The command completed successfully.`
	netUseOutputMultiple = `New connections will be remembered.

Status       Local     Remote                    Network

-------------------------------------------------------------------------------
OK           S:        \\tower\Movies            Microsoft Windows Network
OK                     \\tower\Buildings         Microsoft Windows Network
             T:        \\tower\Music             Microsoft Windows Network
Unavailable  U:        \\tower\Pictures          Microsoft Windows Network
The command completed successfully.`
)

// testCommander implements commander interface.
type testCommander struct {
	allowed []string
}

// Command uses the test executable (taken from os.Args[0]), to execute
// TestNetUseOutput test to emulate `net use` command execution.
func (c testCommander) Command(name string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestNetUseOutput", "--"}
	cs = append(cs, name)
	cs = append(cs, args...)
	// nolint:gosec
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{
		"GO_WANT_TEST_OUTPUT=1",
		fmt.Sprintf("GO_ALLOWED_NET_NAMES=%s", strings.Join(c.allowed, ",")),
	}

	return cmd
}

// TestNetUseOutput is only used to be triggered by testCommander.Command.
// If trigger by testCommander.Command is detected via set GO_WANT_TEST_OUTPUT
// environment variable, it will emulates `net use` command usage by writing
// mocked `net use` output to stdout.
func TestNetUseOutput(*testing.T) {
	if os.Getenv("GO_WANT_TEST_OUTPUT") != "1" {
		return
	}

	args := os.Args

	idx := len(args)
	for i, arg := range args {
		if arg == "--" {
			idx = i + 1
			break
		}
	}

	if idx+1 >= len(args) {
		os.Exit(1)
	}

	name := args[idx]
	if args[idx+1] != "use" {
		os.Exit(1)
	}

	allowed := strings.Split(os.Getenv("GO_ALLOWED_NET_NAMES"), ",")
	isAllowed := false

	for _, candidate := range allowed {
		if candidate == name {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		os.Exit(1)
	}

	fmt.Print(netUseOutputMultiple)
	os.Exit(0)
}

func TestFormatLocalFilePath(t *testing.T) {
	cmd = testCommander{allowed: []string{"net", "net.exe"}}
	formatted, err := FormatLocalFilePath(`X:\localfile`, `S:\entity`)
	require.NoError(t, err)

	assert.Equal(t, `\\tower\Movies\entity`, formatted)
}

func TestFormatLocalFilePath_LocalFileExists(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	cmd = testCommander{allowed: []string{"net", "net.exe"}}
	formatted, err := FormatLocalFilePath(tmpFile.Name(), `S:\entity`)
	require.NoError(t, err)

	assert.Equal(t, tmpFile.Name(), formatted)
}

func TestFormatLocalFilePath_EntityExists(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	cmd = testCommander{allowed: []string{"net", "net.exe"}}
	formatted, err := FormatLocalFilePath(`X:\localfile`, tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, `X:\localfile`, formatted)
}

func TestToUncPath(t *testing.T) {
	cmd = testCommander{allowed: []string{"net", "net.exe"}}
	x, err := toUncPath(`S:\path\to\file`)
	require.NoError(t, err)

	assert.Equal(t, `\\tower\Movies\path\to\file`, x)
}

func TestToUncPath_NoDrive(t *testing.T) {
	cmd = testCommander{allowed: []string{"net", "net.exe"}}
	x, err := toUncPath(`path\to\file`)
	require.NoError(t, err)

	assert.Equal(t, `path\to\file`, x)
}

func TestToUncPath_FallbackToSystem32NetExe(t *testing.T) {
	t.Setenv("WINDIR", `C:\Windows`)

	system32NetExe := filepath.Join(`C:\Windows`, "System32", "net.exe")
	cmd = testCommander{allowed: []string{system32NetExe}}

	x, err := toUncPath(`S:\path\to\file`)
	require.NoError(t, err)

	assert.Equal(t, `\\tower\Movies\path\to\file`, x)
}

func TestParseNetUseOutput(t *testing.T) {
	tests := map[string]struct {
		Output   string
		Expected remoteDrives
	}{
		"one drive": {
			Output: netUseOutputOne,
			Expected: remoteDrives{
				"Z": `\\remotepc\share`,
			},
		},
		"multiple drive": {
			Output: netUseOutputMultiple,
			Expected: remoteDrives{
				"S": `\\tower\Movies`,
				"T": `\\tower\Music`,
				"U": `\\tower\Pictures`,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			drives, err := parseNetUseOutput(test.Output)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, drives)
		})
	}
}

func TestParseNetUseColumns(t *testing.T) {
	columns, err := parseNetUseColumns(`Status       Local     Remote       Network`)
	require.NoError(t, err)

	assert.Equal(t, netUseColumns{
		Local: netUseColumn{
			Start: 13,
			Width: 10,
		},
		Remote: netUseColumn{
			Start: 23,
			Width: 13,
		},
	}, columns)
}

func TestSplitDrive(t *testing.T) {
	tests := map[string]struct {
		Filepath            string
		ExpectedDriveLetter string
		ExpectedPath        string
	}{
		"default": {
			Filepath:            `S:\\remotepc\share`,
			ExpectedDriveLetter: `S`,
			ExpectedPath:        `\\remotepc\share`,
		},
		"lower case drive": {
			Filepath:            `s:\\remotepc\share`,
			ExpectedDriveLetter: `S`,
			ExpectedPath:        `\\remotepc\share`,
		},
		"without drive": {
			Filepath:            `remotepc\share`,
			ExpectedDriveLetter: ``,
			ExpectedPath:        `remotepc\share`,
		},
		"no letter start": {
			Filepath:            `_:\\remotepc\share`,
			ExpectedDriveLetter: ``,
			ExpectedPath:        `_:\\remotepc\share`,
		},
		"one character drive": {
			Filepath:            `A`,
			ExpectedDriveLetter: "",
			ExpectedPath:        `A`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			driveLetter, path := splitDrive(test.Filepath)

			assert.Equal(t, test.ExpectedDriveLetter, driveLetter)
			assert.Equal(t, test.ExpectedPath, path)
		})
	}
}
