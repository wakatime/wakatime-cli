package api

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/mitchellh/go-homedir"
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

// WithHostname sets the X-Machine-Name header to the passed in hostname.
func WithHostname(hostname string) Option {
	return func(c *Client) {
		c.machineNameHeader = hostname
	}
}

// WithDisableSSLVerify disables verification of insecure certificates.
func WithDisableSSLVerify() Option {
	return func(c *Client) {
		var transport *http.Transport
		if c.client.Transport == nil {
			transport = http.DefaultTransport.(*http.Transport).Clone()
		} else {
			transport = c.client.Transport.(*http.Transport).Clone()
		}

		tlsConfig := transport.TLSClientConfig
		tlsConfig.InsecureSkipVerify = true

		transport.TLSClientConfig = tlsConfig
		c.client.Transport = transport
	}
}

// WithSSLCert overrides the default CA certs file to trust specified cert file.
func WithSSLCert(filepath string) (Option, error) {
	expanded, err := homedir.Expand(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed expanding filepath %q: %s", filepath, err)
	}

	caCert, err := ioutil.ReadFile(expanded)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return func(c *Client) {
		var transport *http.Transport
		if c.client.Transport == nil {
			transport = http.DefaultTransport.(*http.Transport).Clone()
		} else {
			transport = c.client.Transport.(*http.Transport).Clone()
		}

		tlsConfig := transport.TLSClientConfig
		tlsConfig.RootCAs = caCertPool
		transport.TLSClientConfig = tlsConfig

		c.client.Transport = transport
	}, nil
}

// WithProxy configures the client to proxy outgoing requests to the specified url.
func WithProxy(proxyURL string) (Option, error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy url %q: %s", proxyURL, err)
	}

	return func(c *Client) {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		if c.client.Transport != nil {
			transport = c.client.Transport.(*http.Transport).Clone()
		}

		transport.Proxy = http.ProxyURL(u)

		c.client.Transport = transport
	}, nil
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
	userAgent := heartbeat.UserAgent(plugin)

	return func(c *Client) {
		c.userAgentHeader = userAgent
	}
}
