package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/wakatime/wakatime-cli/pkg/devsforfile"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

// DevsForFile fetches code stats for Today.
//
// ErrRequest is returned upon request failure with no received response from api.
// ErrAuth is returned upon receiving a 401 Unauthorized api response.
// Err is returned on any other api response related error.
func (c *Client) DevsForFile(heartbeat heartbeat.Heartbeat) (*devsforfile.Devs, error) {
	url := c.baseURL + "/users/current/developers_for_file"

	data, err := json.Marshal(heartbeat)
	if err != nil {
		return nil, fmt.Errorf("failed to json encode body: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, Err{fmt.Errorf("failed to create request: %s", err)}
	}

	req.Header.Set("Content-Type", "application/json")

	q := req.URL.Query()
	req.URL.RawQuery = q.Encode()

	resp, err := c.Do(req)
	if err != nil {
		return nil, Err{fmt.Errorf("failed to make request to %q: %s", url, err)}
	}

	defer resp.Body.Close() // nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, Err{fmt.Errorf("failed to read response body from %q: %s", url, err)}
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusAccepted:
	case http.StatusUnauthorized:
		return nil, ErrAuth{Err: fmt.Errorf("authentication failed at %q. body: %q", url, string(body))}
	case http.StatusBadRequest:
		return nil, ErrBadRequest{fmt.Errorf("bad request at %q", url)}
	default:
		return nil, Err{fmt.Errorf(
			"invalid response status from %q. got: %d, want: %d. body: %q",
			url,
			resp.StatusCode,
			http.StatusOK,
			string(body),
		)}
	}

	summary, err := ParseDevsForFileResponse(body)
	if err != nil {
		return nil, Err{Err: fmt.Errorf("failed to parse results from %q: %s", url, err)}
	}

	return summary, nil
}

// ParseStatusBarResponse parses the wakatime api response into summary.Summary.
func ParseDevsForFileResponse(data []byte) (*devsforfile.Devs, error) {
	var body devsforfile.Devs

	if err := json.Unmarshal(data, &body); err != nil {
		return nil, fmt.Errorf("failed to parse json response body: %s. body: %q", err, data)
	}

	return &body, nil
}
