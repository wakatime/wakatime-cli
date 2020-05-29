package api

import (
	"fmt"
	"runtime"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/version"

	"github.com/matishsiao/goInfo"
)

// Option is a functional option for Client.
type Option func(*Client)

// WithAuth adds authentication via Authorization header.
func WithAuth(auth BasicAuth) (Option, error) {
	authHeaderValue, err := auth.HeaderValue()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve auth header value: %w", err)
	}

	return func(c *Client) {
		c.authHeader = authHeaderValue
	}, nil
}

// WithHostName sets the X-Machine-Name header to the passed in hostname.
func WithHostName(hostname string) Option {
	return func(c *Client) {
		c.machineNameHeader = hostname
	}
}

// WithTimeout configures a timeout for all requests.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.client.Timeout = timeout
	}
}

// WithUserAgentUnknownPlugin sets the User-Agent header on all requests,
// including default value for plugin.
func WithUserAgentUnknownPlugin() Option {
	return WithUserAgent("Unknown/0")
}

// WithUserAgent sets the User-Agent header on all requests, including the passed
// in value for plugin.
func WithUserAgent(plugin string) Option {
	info := goInfo.GetInfo()
	userAgent := fmt.Sprintf(
		"wakatime/%s (%s-%s-%s) %s %s",
		version.Version,
		runtime.GOOS,
		info.Core,
		info.Platform,
		runtime.Version(),
		plugin,
	)

	return func(c *Client) {
		c.userAgentHeader = userAgent
	}
}
