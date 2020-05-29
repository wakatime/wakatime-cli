package api

import "net/http"

// BaseURL is the base url of the wakatime api.
const BaseURL = "https://api.wakatime.com/api"

// Client communicates with the wakatime api.
type Client struct {
	baseURL    string
	client     *http.Client
	userAgent  string
	authHeader string
}

// NewClient creates a new Client. Any number of Options can be provided.
func NewClient(baseURL string, client *http.Client, opts ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		client:  client,
	}

	for _, option := range opts {
		option(c)
	}

	return c
}

// Do wraps c.client.Do() and sets default headers and headers, which are set
// via Option.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Accept", "application/json")

	if c.authHeader != "" {
		req.Header.Set("Authorization", c.authHeader)
	}

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	return c.client.Do(req)
}
