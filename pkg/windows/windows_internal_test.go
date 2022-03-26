package windows

import (
	"fmt"
	"os"
	"os/exec"
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
type testCommander struct{}

// Command uses the test executable (taken from os.Args[0]), to execute
// TestNetUseOutput test to emulate `net use` command execution.
func (testCommander) Command(_ string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestNetUseOutput", "--"}
	cs = append(cs, args...)
	// nolint:gosec
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_TEST_OUTPUT=1"}

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

	defer os.Exit(0)

	fmt.Print(netUseOutputMultiple)
}

func TestFormatLocalFilePath(t *testing.T) {
	cmd = testCommander{}
	formatted, err := FormatLocalFilePath(`X:\localfile`, `S:\entity`)
	require.NoError(t, err)

	assert.Equal(t, `\\tower\Movies\entity`, formatted)
}

func TestFormatLocalFilePath_LocalFileExists(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	cmd = testCommander{}
	formatted, err := FormatLocalFilePath(tmpFile.Name(), `S:\entity`)
	require.NoError(t, err)

	assert.Equal(t, tmpFile.Name(), formatted)
}

func TestFormatLocalFilePath_EntityExists(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	cmd = testCommander{}
	formatted, err := FormatLocalFilePath(`X:\localfile`, tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, `X:\localfile`, formatted)
}

func TestToUncPath(t *testing.T) {
	cmd = testCommander{}
	x, err := toUncPath(`S:\path\to\file`)
	require.NoError(t, err)

	assert.Equal(t, `\\tower\Movies\path\to\file`, x)
}

func TestToUncPath_NoDrive(t *testing.T) {
	cmd = testCommander{}
	x, err := toUncPath(`path\to\file`)
	require.NoError(t, err)

	assert.Equal(t, `path\to\file`, x)
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
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			driveLetter, path := splitDrive(test.Filepath)

			assert.Equal(t, test.ExpectedDriveLetter, driveLetter)
			assert.Equal(t, test.ExpectedPath, path)
		})
	}
}
