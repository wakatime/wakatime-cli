//go:build windows

package api

import (
	"context"
	"crypto/x509"
	"runtime/debug"
	"syscall"
	"unsafe"

	"github.com/wakatime/wakatime-cli/pkg/log"
)

func loadSystemRoots(ctx context.Context) (*x509.CertPool, error) {
	logger := log.Extract(ctx)

	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("panicked: failed to load system roots on Windows: %v. Stack: %s", r, string(debug.Stack()))
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
		if err := syscall.CertCloseStore(store, 0); err != nil {
			logger.Debugf("failed to close system store: %s", err)
		}
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
