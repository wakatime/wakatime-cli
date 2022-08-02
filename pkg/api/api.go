package api

import (
	"errors"
	"net"
	"net/http"

	"github.com/wakatime/wakatime-cli/pkg/log"
)

const (
	// BaseURL is the base url of the wakatime api.
	BaseURL = "https://api.wakatime.com/api/v1"
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
		var dnsError *net.DNSError
		if errors.As(err, &dnsError) {
			log.Warnf("dns error: %s. Retrying with fallback dns resolver", req.URL)

			c.client = &http.Client{
				Transport: NewTransportWithCloudfareDNS(),
			}

			return c.doFunc(c, req)
		}

		return nil, err
	}

	return resp, nil
}
