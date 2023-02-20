package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/wakatime/wakatime-cli/pkg/goal"
)

// Goal fetches goal for the given goal id.
//
// ErrRequest is returned upon request failure with no received response from api.
// ErrAuth is returned upon receiving a 401 Unauthorized api response.
// Err is returned on any other api response related error.
func (c *Client) Goal(id string) (*goal.Goal, error) {
	url := c.baseURL + "/users/current/goals/" + id

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, Err{Err: fmt.Errorf("failed to make request to %q: %s", url, err)}
	}
	defer resp.Body.Close() // nolint:errcheck,gosec

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, Err{Err: fmt.Errorf("failed to read response body from %q: %s", url, err)}
	}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return nil, ErrAuth{Err: fmt.Errorf("authentication failed at %q. body: %q", url, string(body))}
	case http.StatusBadRequest:
		return nil, ErrBadRequest{Err: fmt.Errorf("bad request at %q", url)}
	default:
		return nil, Err{Err: fmt.Errorf(
			"invalid response status from %q. got: %d, want: %d. body: %q",
			url,
			resp.StatusCode,
			http.StatusOK,
			string(body),
		)}
	}

	goal, err := ParseGoalResponse(body)
	if err != nil {
		return nil, Err{Err: fmt.Errorf("failed to parse results from %q: %s", url, err)}
	}

	return goal, nil
}

// ParseGoalResponse parses the wakatime api response into goal.Goal.
func ParseGoalResponse(data []byte) (*goal.Goal, error) {
	var body goal.Goal

	if err := json.Unmarshal(data, &body); err != nil {
		return nil, fmt.Errorf("failed to parse json response body: %s. body: %q", err, data)
	}

	return &body, nil
}
