package api

import (
	"net/http"
	"time"
)

// BaseURL is the base url of the wakatime api.
const BaseURL = "https://api.wakatime.com/api/v1"

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
			Transport:     NewTransport(),
			CheckRedirect: nil, // defaults to following up to 10 redirects
			Timeout:       30 * time.Second,
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
	return c.doFunc(c, req)
}
