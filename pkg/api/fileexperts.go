package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/wakatime/wakatime-cli/pkg/fileexperts"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// FileExperts fetches file experts for Today.
//
// ErrRequest is returned upon request failure with no received response from api.
// ErrAuth is returned upon receiving a 401 Unauthorized api response.
// Err is returned on any other api response related error.
func (c *Client) FileExperts(heartbeats []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
	url := c.baseURL + "/users/current/file_experts"

	// change from heartbeat.Heartbeat to fileexpert.Entity
	// it's safe to get the first item in the slice.
	e := fileexperts.Entity{
		Filepath:         heartbeats[0].Entity,
		Project:          heartbeats[0].Project,
		ProjectRootCount: heartbeats[0].ProjectRootCount,
	}

	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("failed to json encode body: %s", err)
	}

	log.Debugf("fileexpert: %s", string(data))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// set auth header here for every request due to multiple api key support
	setAuthHeader(req, heartbeats[0].APIKey)

	resp, err := c.Do(req)
	if err != nil {
		return nil, Err{Err: fmt.Errorf("failed making request to %q: %s", url, err)}
	}
	defer resp.Body.Close() // nolint:errcheck,gosec

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, Err{Err: fmt.Errorf("failed reading response body from %q: %s", url, err)}
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

	results, err := ParseFileExpertsResponse(body)
	if err != nil {
		return nil, Err{Err: fmt.Errorf("failed to parse results from %q: %s", url, err)}
	}

	return results, nil
}

// ParseFileExpertsResponse parses the wakatime api response into fileexperts.FileExperts.
func ParseFileExpertsResponse(data []byte) ([]heartbeat.Result, error) {
	var body fileexperts.FileExperts

	if err := json.Unmarshal(data, &body); err != nil {
		return nil, fmt.Errorf("failed to parse json response body: %s. body: %q", err, data)
	}

	return []heartbeat.Result{{FileExpert: &body}}, nil
}
