// +build !windows

package api

import (
	"crypto/x509"
)

func loadSystemRoots() (*x509.CertPool, error) {
	return x509.SystemCertPool()
}
