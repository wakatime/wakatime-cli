package windows

import (
	"fmt"
	"io/ioutil"
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
             Z:        \\remotepc\share          Microsoft Windows Network
The command completed successfully.`
	netUseOutputMultiple = `New connections will be remembered.

Status       Local     Remote                    Network

-------------------------------------------------------------------------------
             S:        \\tower\Movies            Microsoft Windows Network
             T:        \\tower\Music             Microsoft Windows Network
             U:        \\tower\Pictures          Microsoft Windows Network
The command completed successfully.`
)

type testCommander struct{}

func (c testCommander) Command(name string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestOutput", "--"}
	cs = append(cs, args...)
	// nolint:gosec
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_TEST_OUTPUT=1"}

	return cmd
}

func TestOutput(*testing.T) {
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
	tmpFile, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	cmd = testCommander{}
	formatted, err := FormatLocalFilePath(tmpFile.Name(), `S:\entity`)
	require.NoError(t, err)

	assert.Equal(t, tmpFile.Name(), formatted)
}

func TestFormatLocalFilePath_EntityExists(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

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
