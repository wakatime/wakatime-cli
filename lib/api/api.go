package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Config struct {
	APIKey    string
	HostName  string
	UserAgent string
}

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string, client *http.Client) *Client {
	return &Client{
		baseURL: baseURL,
		client:  client,
	}
}

func (c *Client) SendHeartbeats(heartbeats []Heartbeat, cfg Config) error {
	url := c.baseURL + "/users/current/heartbeats.bulk"

	data, err := json.Marshal(heartbeats)
	if err != nil {
		return fmt.Errorf("failed to json encode body: %s", err)
	}
	log.Printf("Sending heartbeats to api at %q. request body: %s", url, string(data))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %s", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(cfg.APIKey))))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", cfg.UserAgent)
	if cfg.HostName != "" {
		req.Header.Set("X-Machine-Name", cfg.HostName)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return NewErr("failed making request to %q: %s", url, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return NewErr("failed reading response body from %q: %s", url, err)
	}

	switch resp.StatusCode {
	case http.StatusCreated, http.StatusAccepted:
		break
	case http.StatusUnauthorized:
		return NewErrAuth("authentication failed at %q", url)
	default:
		return NewErr(
			"invalid response status from %q. got: %d, want: %d/%d. body: %q",
			url,
			resp.StatusCode,
			http.StatusCreated,
			http.StatusAccepted,
			string(body),
		)
	}

	return nil
}
