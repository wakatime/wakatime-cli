package api

import (
	"net/http"
	"time"
)

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
