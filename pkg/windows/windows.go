package windows

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"unicode"
)

// nolint
var (
	backslashReplaceRegex = regexp.MustCompile(`[\\/]+`)
	ipv4seg               = "(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])"
	ipv4Address           = fmt.Sprintf(`(%s\.){3,3}%s`, ipv4seg, ipv4seg)
	ipv6seg               = "[0-9a-fA-F]{1,4}"
	ipv6Address           = fmt.Sprintf("("+
		"(%s:){7,7}%s|"+
		"(%s:){1,7}:|"+
		"(%s:){1,6}:%s|"+
		"(%s:){1,5}(:%s){1,2}|"+
		"(%s:){1,4}(:%s){1,3}|"+
		"(%s:){1,3}(:%s){1,4}|"+
		"(%s:){1,2}(:%s){1,5}|"+
		"%s:((:%s){1,6})|"+
		":((:%s){1,7}|:)|"+
		"fe80:(:%s){0,4}%%[0-9a-zA-Z]{1,}|"+
		"::(ffff(:0{1,4}){0,1}:){0,1}%s|"+
		"(%s:){1,4}:%s)", ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg,
		ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg,
		ipv6seg, ipv4Address, ipv6seg, ipv4Address)
	windowsDriveRegex        = regexp.MustCompile("^[a-z]:/")
	windowsNetworkMountRegex = regexp.MustCompile(fmt.Sprintf(`(?i)^\\\\([a-z]|%s|%s)+`, ipv4Address, ipv6Address))
)

// FormatFilePath formats a windows filepath by converting backslash to
// frontslash and ensuring that drive letter is upper case.
func FormatFilePath(fp string) (string, error) {
	isWindowsNetworkMount := windowsNetworkMountRegex.MatchString(fp)

	fp = backslashReplaceRegex.ReplaceAllString(fp, "/")

	if windowsDriveRegex.MatchString(fp) {
		fp = strings.ToUpper(fp[:1]) + fp[1:]
	}

	if isWindowsNetworkMount {
		// Add back a / to the front, since the previous modifications
		// will have replaced any double slashes with single
		fp = "/" + fp
	}

	return fp, nil
}

// IsWindowsNetworkMount returns true if filepath is windows network path.
func IsWindowsNetworkMount(fp string) bool {
	return windowsNetworkMountRegex.MatchString(fp)
}

// commander is an interface for exec.Command function.
type commander interface {
	Command(name string, args ...string) *exec.Cmd
}

// realCommander implements commander interface and is used by default.
type realCommander struct{}

// Command calls exec.Command function.
func (c realCommander) Command(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// nolint:gochecknoglobals
// commander replaces exec.Command function. It is initialized in init()
// and can be overwritten in tests.
var cmd commander

// nolint:gochecknoinits
func init() {
	cmd = realCommander{}
}

// FormatLocalFilePath maps entity filepath to unc path, if neither
// localFile, nor entity file are existing.
func FormatLocalFilePath(localFile, entity string) (string, error) {
	// if entity exists, do nothing
	if info, err := os.Stat(entity); err == nil && !info.IsDir() {
		return localFile, nil
	}

	// if local file exists, do nothing
	if info, err := os.Stat(localFile); err == nil && !info.IsDir() {
		return localFile, nil
	}

	uncPath, err := toUncPath(entity)
	if err != nil {
		return "", fmt.Errorf("failed to convert entity %q to unc path: %s", entity, err)
	}

	return uncPath, nil
}

// toUncPath converts a filepath to a Universal Naming Convention path
// by querying remote drive information via `net use` cmd.
func toUncPath(fp string) (string, error) {
	letter, rest := splitDrive(fp)
	if letter == "" {
		return fp, nil
	}

	out, err := cmd.Command("net use").Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute ls command: %s", err)
	}

	drives, err := parseNetUseOutput(string(out))
	if err != nil {
		return "", fmt.Errorf("failed to parse net use command output: %s", err)
	}

	if drive, ok := drives[driveLetter(letter)]; ok {
		return string(drive) + rest, nil
	}

	return fp, nil
}

// driveLetter represents the letter of a drive.
type driveLetter string

// remoteDrive represents the path to a remote drive.
type remoteDrive string

// remoteDrives maps drive letters to remote drives.
type remoteDrives map[driveLetter]remoteDrive

// parseNetUseOutput parses the drives from net use output.
func parseNetUseOutput(text string) (remoteDrives, error) {
	var (
		cols netUseColumns
		err  error
	)

	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")

	drives := make(remoteDrives)

	for _, line := range lines[1 : len(lines)-1] {
		if len(strings.TrimSpace(line)) == 0 || strings.ContainsAny(line, "---") {
			continue
		}

		if cols.Empty() {
			cols, err = parseNetUseColumns(line)
			if err != nil {
				return nil, fmt.Errorf("failed to parse columns from 'net use' output: %s", err)
			}

			continue
		}

		local := line[cols.Local.Start : cols.Local.Start+cols.Remote.Width]
		local = strings.ToUpper(strings.TrimSpace(local))
		letter := strings.Split(local, ":")[0][0]

		if !unicode.IsLetter(rune(letter)) {
			continue
		}

		remote := line[cols.Remote.Start : cols.Remote.Start+cols.Remote.Width]

		drives[driveLetter(letter)] = remoteDrive(strings.TrimSpace(remote))
	}

	return drives, nil
}

// netUseColumn represents a column of the 'net use' windows command output.
// It has a start and end position and is used to parse the listed mapped
// network drives.
type netUseColumn struct {
	Start int
	Width int
}

// Empty returns true, if netUseColumn is unset.
func (c netUseColumn) Empty() bool {
	if c.Start == 0 && c.Width == 0 {
		return true
	}

	return false
}

// netUseColumn represents the column of the 'net use' windows command output.
// Only the local and remote column are of importance here.
type netUseColumns struct {
	Local  netUseColumn
	Remote netUseColumn
}

// Empty returns true, if all netUseColumns are unset.
func (c netUseColumns) Empty() bool {
	return c.Local.Empty() && c.Remote.Empty()
}

// parseNetUseColumns parses the column line of the 'net use' windows command
// to determine their start position and width.
func parseNetUseColumns(line string) (netUseColumns, error) {
	re := regexp.MustCompile(`[a-zA-Z]+[^a-zA-Z]*`)
	matches := re.FindAllString(line, -1)

	var (
		cols  netUseColumns
		start int
	)

	for _, match := range matches {
		key := strings.ToLower(strings.TrimSpace(match))

		switch key {
		case "local":
			cols.Local = netUseColumn{
				Start: start,
				Width: len(match),
			}
		case "remote":
			cols.Remote = netUseColumn{
				Start: start,
				Width: len(match),
			}
		}

		start += len(match)
	}

	if cols.Local.Empty() {
		return netUseColumns{}, errors.New("failed to parse local column")
	} else if cols.Remote.Empty() {
		return netUseColumns{}, errors.New("failed to parse remote column")
	}

	return cols, nil
}

// splitDrive splits a filepath into the drive letter and the path.
func splitDrive(fp string) (string, string) {
	if fp[1:2] != ":" || !unicode.IsLetter(rune(fp[0])) {
		return "", fp
	}

	return strings.ToUpper(string(fp[0])), fp[2:]
}
