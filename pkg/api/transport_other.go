//go:build !windows

package api

import (
	"context"
	"crypto/x509"
)

func loadSystemRoots(_ context.Context) (*x509.CertPool, error) {
	return x509.SystemCertPool()
}
