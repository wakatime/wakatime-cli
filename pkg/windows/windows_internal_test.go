package windows

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatLocalFilePath(t *testing.T) {
	restore := stubWindowsAPIs(
		func(string) (uint32, error) { return driveRemote, nil },
		func(string) (string, error) { return `\\tower\Movies\entity`, nil },
		func(string) (string, error) { return `\\tower\Movies\entity`, nil },
	)
	defer restore()

	formatted, err := FormatLocalFilePath(`X:\localfile`, `S:\entity`)
	require.NoError(t, err)

	assert.Equal(t, `\\tower\Movies\entity`, formatted)
}

func TestFormatLocalFilePath_LocalFileExists(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	restore := stubWindowsAPIs(
		func(string) (uint32, error) { return driveRemote, nil },
		func(string) (string, error) { return `\\tower\Movies\entity`, nil },
		func(string) (string, error) { return `\\tower\Movies\entity`, nil },
	)
	defer restore()

	formatted, err := FormatLocalFilePath(tmpFile.Name(), `S:\entity`)
	require.NoError(t, err)

	assert.Equal(t, tmpFile.Name(), formatted)
}

func TestFormatLocalFilePath_EntityExists(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	restore := stubWindowsAPIs(
		func(string) (uint32, error) { return driveRemote, nil },
		func(string) (string, error) { return `\\tower\Movies\entity`, nil },
		func(string) (string, error) { return `\\tower\Movies\entity`, nil },
	)
	defer restore()

	formatted, err := FormatLocalFilePath(`X:\localfile`, tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, `X:\localfile`, formatted)
}

func TestToUncPath(t *testing.T) {
	restore := stubWindowsAPIs(
		func(root string) (uint32, error) {
			assert.Equal(t, `S:\`, root)
			return driveRemote, nil
		},
		func(path string) (string, error) {
			assert.Equal(t, `S:\path\to\file`, path)
			return `\\tower\Movies\path\to\file`, nil
		},
		func(string) (string, error) {
			t.Fatal("WNetGetConnection fallback should not be used for full path")
			return "", nil
		},
	)
	defer restore()

	x, err := toUncPath(`S:\path\to\file`)
	require.NoError(t, err)

	assert.Equal(t, `\\tower\Movies\path\to\file`, x)
}

func TestToUncPath_NoDrive(t *testing.T) {
	x, err := toUncPath(`path\to\file`)
	require.NoError(t, err)

	assert.Equal(t, `path\to\file`, x)
}

func TestToUncPath_LocalDrive(t *testing.T) {
	restore := stubWindowsAPIs(
		func(root string) (uint32, error) {
			assert.Equal(t, `C:\`, root)
			return 3, nil
		},
		func(string) (string, error) {
			t.Fatal("WNetGetUniversalName should not be called for local drives")
			return "", nil
		},
		func(string) (string, error) {
			t.Fatal("WNetGetConnection should not be called for local drives")
			return "", nil
		},
	)
	defer restore()

	x, err := toUncPath(`C:\path\to\file`)
	require.NoError(t, err)

	assert.Equal(t, `C:\path\to\file`, x)
}

func TestToUncPath_FallbackToConnectionName(t *testing.T) {
	restore := stubWindowsAPIs(
		func(string) (uint32, error) { return driveRemote, nil },
		func(path string) (string, error) {
			assert.Equal(t, `S:\path\to\file`, path)
			return "", errNotSupported
		},
		func(localName string) (string, error) {
			assert.Equal(t, "S:", localName)
			return `\\tower\Movies`, nil
		},
	)
	defer restore()

	x, err := toUncPath(`S:\path\to\file`)
	require.NoError(t, err)

	assert.Equal(t, `\\tower\Movies\path\to\file`, x)
}

func TestToUncPath_RootPathUsesConnectionName(t *testing.T) {
	restore := stubWindowsAPIs(
		func(string) (uint32, error) { return driveRemote, nil },
		func(string) (string, error) {
			t.Fatal("WNetGetUniversalName should not be called for bare drive roots")
			return "", nil
		},
		func(localName string) (string, error) {
			assert.Equal(t, "S:", localName)
			return `\\tower\Movies`, nil
		},
	)
	defer restore()

	x, err := toUncPath(`S:`)
	require.NoError(t, err)

	assert.Equal(t, `\\tower\Movies`, x)
}

func TestToUncPath_GetDriveTypeError(t *testing.T) {
	restore := stubWindowsAPIs(
		func(string) (uint32, error) { return 0, errors.New("boom") },
		func(string) (string, error) { return "", nil },
		func(string) (string, error) { return "", nil },
	)
	defer restore()

	_, err := toUncPath(`S:\path\to\file`)
	require.Error(t, err)
	assert.ErrorContains(t, err, `failed to get drive type for "S:\\"`)
}

func TestToUncPath_GetUniversalNameError(t *testing.T) {
	restore := stubWindowsAPIs(
		func(string) (uint32, error) { return driveRemote, nil },
		func(string) (string, error) { return "", errors.New("boom") },
		func(string) (string, error) { return "", nil },
	)
	defer restore()

	_, err := toUncPath(`S:\path\to\file`)
	require.Error(t, err)
	assert.ErrorContains(t, err, `failed to get universal path for "S:\\path\\to\\file"`)
}

func TestToUncPath_GetConnectionNameError(t *testing.T) {
	restore := stubWindowsAPIs(
		func(string) (uint32, error) { return driveRemote, nil },
		func(string) (string, error) { return "", errNotSupported },
		func(string) (string, error) { return "", errors.New("boom") },
	)
	defer restore()

	_, err := toUncPath(`S:\path\to\file`)
	require.Error(t, err)
	assert.ErrorContains(t, err, `failed to get connection name for "S:"`)
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

func stubWindowsAPIs(
	driveType func(string) (uint32, error),
	universalName func(string) (string, error),
	connectionName func(string) (string, error),
) func() {
	prevDriveType := getDriveType
	prevUniversalName := getUniversalName
	prevConnectionName := getConnectionName

	getDriveType = driveType
	getUniversalName = universalName
	getConnectionName = connectionName

	return func() {
		getDriveType = prevDriveType
		getUniversalName = prevUniversalName
		getConnectionName = prevConnectionName
	}
}
