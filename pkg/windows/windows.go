package windows

import (
	"errors"
	"fmt"
	"os"
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
		"(%s:){7,7}%s|"+ // 1:2:3:4:5:6:7:8
		"(%s:){1,7}:|"+ // 1:: or 1:2:3:4:5:6:7::
		"(%s:){1,6}:%s|"+ // 1::8 or 1:2:3:4:5:6::8 or 1:2:3:4:5:6::8
		"(%s:){1,5}(:%s){1,2}|"+ // 1::7:8 or 1:2:3:4:5::7:8 or 1:2:3:4:5::8
		"(%s:){1,4}(:%s){1,3}|"+ // 1::6:7:8 or 1:2:3:4::6:7:8 or 1:2:3:4::8
		"(%s:){1,3}(:%s){1,4}|"+ // 1::5:6:7:8 or 1:2:3::5:6:7:8 or 1:2:3::8
		"(%s:){1,2}(:%s){1,5}|"+ // 1::4:5:6:7:8 or 1:2::4:5:6:7:8 or 1:2::8
		"%s:((:%s){1,6})|"+ // 1::3:4:5:6:7:8 or 1::3:4:5:6:7:8 or 1::8
		":((:%s){1,7}|:)|"+ // ::2:3:4:5:6:7:8 or ::2:3:4:5:6:7:8 or ::8 or ::
		"fe80:(:%s){0,4}%%[0-9a-zA-Z]{1,}|"+ // fe80::7:8%eth0 or fe80::7:8%1 (link-local IPv6 addresses with zone index)
		"::(ffff(:0{1,4}){0,1}:){0,1}%s|"+ // ::255.255.255.255 or ::ffff:255.255.255.255 or ::ffff:0:255.255.255.255 (IPv4-mapped IPv6 addresses and IPv4-translated addresses)
		"(%s:){1,4}:%s)", // 2001:db8:3:4::192.0.2.33 or 64:ff9b::192.0.2.33 (IPv4-Embedded IPv6 Address)
		ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg,
		ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv6seg, ipv4Address,
		ipv6seg, ipv4Address)
	windowsDriveRegex        = regexp.MustCompile("^[/\\]?[a-z]:/")
	windowsNetworkMountRegex = regexp.MustCompile(fmt.Sprintf(`(?i)^\\\\([a-z]|%s|%s)+`, ipv4Address, ipv6Address))
)

// FormatFilePath formats a windows filepath by converting backslash to
// frontslash and ensuring that drive letter is upper case.
func FormatFilePath(fp string) string {
	isWindowsNetworkMount := IsWindowsNetworkMount(fp)

	fp = backslashReplaceRegex.ReplaceAllString(fp, "/")

	if isWindowsNetworkMount {
		// Replace the first single slash with double backslash, since the regex
		// will have replaced any double backslashes with single forward slash
		fp = `\\` + fp[1:]
	}

	if windowsDriveRegex.MatchString(fp) {
		fp = strings.ToUpper(fp[:1]) + fp[1:]
	}

	return fp
}

// IsWindowsNetworkMount returns true if filepath is windows network path.
func IsWindowsNetworkMount(fp string) bool {
	return windowsNetworkMountRegex.MatchString(fp)
}

const driveRemote = 4

// nolint:gochecknoglobals // Package-level seams for stubbing Win32 API calls in tests.
var (
	getDriveType      = systemGetDriveType
	getUniversalName  = systemGetUniversalName
	getConnectionName = systemGetConnectionName
)

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
// by querying Windows networking APIs for mapped drive information.
func toUncPath(fp string) (string, error) {
	letter, rest := splitDrive(fp)
	if letter == "" {
		return fp, nil
	}

	drive := letter + `:\`

	driveType, err := getDriveType(drive)
	if err != nil {
		return "", fmt.Errorf("failed to get drive type for %q: %w", drive, err)
	}

	if driveType != driveRemote {
		return fp, nil
	}

	if rest != "" {
		uncPath, err := getUniversalName(fp)
		if err == nil {
			return uncPath, nil
		}

		if !errors.Is(err, errNotSupported) {
			return "", fmt.Errorf("failed to get universal path for %q: %w", fp, err)
		}
	}

	connection, err := getConnectionName(letter + ":")
	if err != nil {
		return "", fmt.Errorf("failed to get connection name for %q: %w", letter+":", err)
	}

	return connection + rest, nil
}

// splitDrive splits a filepath into the drive letter and the path.
func splitDrive(fp string) (string, string) {
	if len(fp) < 2 || fp[1:2] != ":" || !unicode.IsLetter(rune(fp[0])) {
		return "", fp
	}

	return strings.ToUpper(string(fp[0])), fp[2:]
}
