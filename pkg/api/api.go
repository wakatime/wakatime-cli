package api

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/log"
)

const (
	// BaseURL is the base url of the wakatime api.
	BaseURL = "https://api.wakatime.com/api/v1"
	// baseIPAddrv4 is the base ip address v4 of the wakatime api.
	baseIPAddrv4 = "143.244.210.202"
	// baseIPAddrv6 is the base ip address v6 of the wakatime api.
	baseIPAddrv6 = "2604:a880:4:1d0::2a7:b000"
	// DefaultTimeoutSecs is the default timeout used for requests to the wakatime api.
	DefaultTimeoutSecs = 120
)

// Client communicates with the wakatime api.
type Client struct {
	baseURL string
	client  *http.Client
	// doFunc allows api client options to manipulate request/response handling.
	// default function will be set in constructor.
	//
	// wrapping by api options should be performed as follows:
	//
	//	next := c.doFunc
	//	c.doFunc = func(c *Client, req *http.Request) (*http.Response, error) {
	//		// do something
	//		resp, err := next(c, req)
	//		// do more
	//		return resp, err
	//	}
	doFunc func(c *Client, req *http.Request) (*http.Response, error)
}

// NewClient creates a new Client. Any number of Options can be provided.
func NewClient(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		client: &http.Client{
			Transport: NewTransport(),
		},
		doFunc: func(c *Client, req *http.Request) (*http.Response, error) {
			req.Header.Set("Accept", "application/json")
			return c.client.Do(req)
		},
	}

	for _, option := range opts {
		option(c)
	}

	return c
}

// Do executes c.doFunc(), which in turn allows wrapping c.client.Do() and manipulating
// the request behavior of the api client.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.doFunc(c, req)
	if err != nil {
		// don't set alternate host if there's a custom api url
		if !strings.HasPrefix(c.baseURL, BaseURL) {
			return nil, err
		}

		var dnsError *net.DNSError
		if !errors.As(err, &dnsError) {
			return nil, err
		}

		c.client = &http.Client{
			Transport: NewTransportWithHostVerificationDisabled(),
		}

		req.URL.Host = baseIPAddrv4
		if isLocalIPv6() {
			req.URL.Host = baseIPAddrv6
		}

		log.Debugf("dns error, will retry with host ip '%s': %s", req.URL.Host, err)

		resp, errRetry := c.doFunc(c, req)
		if errRetry != nil {
			return nil, fmt.Errorf("retry request failed: %s. original error: %s", errRetry, err)
		}

		return resp, nil
	}

	return resp, nil
}

func isLocalIPv6() bool {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:80", baseIPAddrv4))
	if err != nil {
		log.Warnf("failed dialing to detect default local ip address: %s", err)
		return true
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Debugf("failed to close connection to api wakatime: %s", err)
		}
	}()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.To4() == nil
}
