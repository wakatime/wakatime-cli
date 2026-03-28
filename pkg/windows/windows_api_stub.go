//go:build !windows

package windows

import "errors"

var errNotSupported = errors.New("operation not supported")

func systemGetDriveType(string) (uint32, error) {
	return 0, errNotSupported
}

func systemGetConnectionName(string) (string, error) {
	return "", errNotSupported
}

func systemGetUniversalName(string) (string, error) {
	return "", errNotSupported
}
