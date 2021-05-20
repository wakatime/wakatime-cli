package api

import (
	"net/http"
	"time"
)

// NewTransport creates a new http.Transport, with default 30 second TLS timeout.
func NewTransport() *http.Transport {
	return &http.Transport{
		Proxy:               nil,
		TLSHandshakeTimeout: 30 * time.Second,
		MaxIdleConns:        1,
		MaxIdleConnsPerHost: 1,
		MaxConnsPerHost:     1,
		ForceAttemptHTTP2:   true,
	}
}
