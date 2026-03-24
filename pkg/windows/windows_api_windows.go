//go:build windows

package windows

import (
	"errors"
	"fmt"
	"unsafe"

	xsyswindows "golang.org/x/sys/windows"
)

var errNotSupported = errors.New("operation not supported")

const (
	universalNameInfoLevel = 0x00000001
	errorMoreData          = 234
)

type universalNameInfo struct {
	UniversalName *uint16
}

func wNetGetConnectionProc() *xsyswindows.LazyProc {
	return xsyswindows.NewLazySystemDLL("mpr.dll").NewProc("WNetGetConnectionW")
}

func wNetGetUniversalNameProc() *xsyswindows.LazyProc {
	return xsyswindows.NewLazySystemDLL("mpr.dll").NewProc("WNetGetUniversalNameW")
}

//nolint:gosec // Win32 syscall boundaries require explicit pointer-to-uintptr conversion.
func uintptrFrom[T any](value *T) uintptr {
	return uintptr(unsafe.Pointer(value))
}

func systemGetDriveType(rootPath string) (uint32, error) {
	path, err := xsyswindows.UTF16PtrFromString(rootPath)
	if err != nil {
		return 0, err
	}

	driveType := xsyswindows.GetDriveType(path)
	if driveType == 0 {
		return 0, errors.New("GetDriveTypeW returned DRIVE_UNKNOWN")
	}

	return driveType, nil
}

func systemGetConnectionName(localName string) (string, error) {
	name, err := xsyswindows.UTF16PtrFromString(localName)
	if err != nil {
		return "", err
	}

	size := uint32(260)
	for {
		buffer := make([]uint16, size)
		ret, _, _ := wNetGetConnectionProc().Call(
			uintptrFrom(name),
			uintptrFrom(&buffer[0]),
			uintptrFrom(&size),
		)

		switch ret {
		case 0:
			return xsyswindows.UTF16ToString(buffer), nil
		case errorMoreData:
			continue
		case 50:
			return "", errNotSupported
		default:
			return "", fmt.Errorf("WNetGetConnectionW returned %d", ret)
		}
	}
}

func systemGetUniversalName(path string) (string, error) {
	name, err := xsyswindows.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}

	size := uint32(1024)
	for {
		buffer := make([]byte, size)
		ret, _, _ := wNetGetUniversalNameProc().Call(
			uintptrFrom(name),
			uintptr(universalNameInfoLevel),
			uintptrFrom(&buffer[0]),
			uintptrFrom(&size),
		)

		switch ret {
		case 0:
			//nolint:gosec // The buffer is populated by WNetGetUniversalNameW with a UNIVERSAL_NAME_INFOW header.
			info := (*universalNameInfo)(unsafe.Pointer(&buffer[0]))
			return xsyswindows.UTF16PtrToString(info.UniversalName), nil
		case errorMoreData:
			continue
		case 50, 1200:
			return "", errNotSupported
		default:
			return "", fmt.Errorf("WNetGetUniversalNameW returned %d", ret)
		}
	}
}
