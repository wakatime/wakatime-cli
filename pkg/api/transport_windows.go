//go:build windows

package api

import (
	"crypto/x509"
	"runtime/debug"
	"syscall"
	"unsafe"

	"github.com/wakatime/wakatime-cli/pkg/log"
)

func loadSystemRoots() (*x509.CertPool, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("failed to load system roots on Windows. panicked: %v. Stack: %s", err, string(debug.Stack()))
		}
	}()

	const cryptENotFound = 0x80092004

	rootPtr, err := syscall.UTF16PtrFromString("ROOT")
	if err != nil {
		return nil, err
	}

	store, err := syscall.CertOpenSystemStore(0, rootPtr)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = syscall.CertCloseStore(store, 0)
	}()

	roots := x509.NewCertPool()

	var cert *syscall.CertContext

	for {
		cert, err = syscall.CertEnumCertificatesInStore(store, cert)
		if err != nil {
			if errno, ok := err.(syscall.Errno); ok {
				if errno == cryptENotFound {
					break
				}
			}

			return nil, err
		}

		if cert == nil {
			break
		}
		// Copy the buf, since ParseCertificate does not create its own copy.
		buf := (*[1 << 20]byte)(unsafe.Pointer(cert.EncodedCert))[:cert.Length:cert.Length] // nolint:gosec
		buf2 := make([]byte, cert.Length)
		copy(buf2, buf)

		if c, err := x509.ParseCertificate(buf2); err == nil {
			roots.AddCert(c)
		}
	}

	return roots, nil
}
