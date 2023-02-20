package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// SendHeartbeats sends a bulk of heartbeats to the wakatime api and returns the result.
// The API does not guarantuee the setting of the Heartbeat property of the result.
// On certain errors, like 429/too many heartbeats, this is omitted and not set.
//
// ErrRequest is returned upon request failure with no received response from api.
// ErrAuth is returned upon receiving a 401 Unauthorized api response.
// Err is returned on any other api response related error.
func (c *Client) SendHeartbeats(heartbeats []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
	url := c.baseURL + "/users/current/heartbeats.bulk"

	log.Debugf("sending %d heartbeat(s) to api at %s", len(heartbeats), url)

	var results []heartbeat.Result

	grouped := groupByAPIKey(heartbeats)
	keys := sortKeys(grouped)

	for _, k := range keys {
		res, err := c.sendHeartbeats(url, grouped[k])
		if err != nil {
			return nil, err
		}

		results = append(results, res...)
	}

	return results, nil
}

func (c *Client) sendHeartbeats(url string, heartbeats []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
	data, err := json.Marshal(heartbeats)
	if err != nil {
		return nil, fmt.Errorf("failed to json encode body: %s", err)
	}

	log.Debugf("heartbeats: %s", string(data))

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
	defer resp.Body.Close() // nolint:errcheck,gosec,gosec

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, Err{Err: fmt.Errorf("failed reading response body from %q: %s", url, err)}
	}

	switch resp.StatusCode {
	case http.StatusCreated, http.StatusAccepted:
	case http.StatusUnauthorized:
		return nil, ErrAuth{Err: fmt.Errorf("authentication failed at %q", url)}
	case http.StatusBadRequest:
		return nil, ErrBadRequest{Err: fmt.Errorf("bad request at %q", url)}
	default:
		return nil, Err{Err: fmt.Errorf(
			"invalid response status from %q. got: %d, want: %d/%d. body: %q",
			url,
			resp.StatusCode,
			http.StatusCreated,
			http.StatusAccepted,
			string(body),
		)}
	}

	results, err := ParseHeartbeatResponses(body)
	if err != nil {
		return nil, Err{Err: fmt.Errorf("failed parsing results from %q: %s", url, err)}
	}

	return results, nil
}

// ParseHeartbeatResponses parses the aggregated responses returned by the heartbeat bulk endpoint.
func ParseHeartbeatResponses(data []byte) ([]heartbeat.Result, error) {
	var responsesBody struct {
		Responses [][]json.RawMessage `json:"responses"`
	}

	err := json.Unmarshal(data, &responsesBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json response body: %s. body: %q", err, string(data))
	}

	var results []heartbeat.Result

	for n, r := range responsesBody.Responses {
		result, err := parseHeartbeatResponse(r)
		if err != nil {
			return nil, fmt.Errorf("failed parsing result #%d: %s. body: %q", n, err, string(data))
		}

		results = append(results, result)
	}

	return results, nil
}

// parseHeartbeatResponse parses one response of the aggregated responses returned by the heartbeat bulk endpoint.
func parseHeartbeatResponse(data []json.RawMessage) (heartbeat.Result, error) {
	var result heartbeat.Result

	type responseBody struct {
		Data *heartbeat.Heartbeat `json:"data"`
	}

	err := json.Unmarshal(data[1], &result.Status)
	if err != nil {
		return heartbeat.Result{}, fmt.Errorf("failed to parse json status: %s", err)
	}

	if result.Status >= http.StatusBadRequest {
		resultErrors, err := parseHeartbeatResponseError(data[0])
		if err != nil {
			return heartbeat.Result{}, fmt.Errorf("failed to parse result errors: %s", err)
		}

		result.Errors = resultErrors

		return heartbeat.Result{
			Errors: result.Errors,
			Status: result.Status,
		}, nil
	}

	err = json.Unmarshal(data[0], &responseBody{Data: &result.Heartbeat})
	if err != nil {
		return heartbeat.Result{}, fmt.Errorf("failed to parse json heartbeat: %s", err)
	}

	return result, nil
}

// parseHeartbeatResponseError parses one error of the aggregated responses returned by the heartbeat bulk endpoint.
func parseHeartbeatResponseError(data json.RawMessage) ([]string, error) {
	var errs []string

	type responseBodyErr struct {
		Error  *string         `json:"error"`
		Errors *map[string]any `json:"errors"`
	}

	// 1. try "error" key
	var resultError string

	err := json.Unmarshal(data, &responseBodyErr{Error: &resultError})
	if err != nil {
		log.Debugf("failed to parse json heartbeat error or 'error' key not found: %s", err)
	}

	if resultError != "" {
		errs = append(errs, resultError)
		return errs, nil
	}

	// 2. try "errors" key
	var resultErrors map[string]any

	err = json.Unmarshal(data, &responseBodyErr{Errors: &resultErrors})
	if err != nil {
		log.Debugf("failed to parse json heartbeat errors or 'errors' key not found: %s", err)
	}

	if resultErrors == nil {
		return nil, errors.New("failed to detect any errors despite invalid response status")
	}

	for field, messages := range resultErrors {
		// skipping parsing dependencies errors as it won't happen because we are
		// filtering in the cli.
		if field == "dependencies" {
			continue
		}

		m := make([]string, len(messages.([]any)))
		for i, v := range messages.([]any) {
			m[i] = fmt.Sprint(v)
		}

		errs = append(errs, fmt.Sprintf(
			"%s: %s",
			field,
			strings.Join(m, " "),
		))
	}

	return errs, nil
}

func groupByAPIKey(hh []heartbeat.Heartbeat) map[string][]heartbeat.Heartbeat {
	var grouped = make(map[string][]heartbeat.Heartbeat, 0)

	for _, h := range hh {
		grouped[h.APIKey] = append(grouped[h.APIKey], h)
	}

	return grouped
}

func sortKeys[K string, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	i := 0

	for k := range m {
		keys[i] = k
		i++
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	return keys
}

func setAuthHeader(req *http.Request, apiKey string) {
	authHeaderValue, _ := BasicAuth{Secret: apiKey}.HeaderValue()

	req.Header.Set("Authorization", authHeaderValue)
}
