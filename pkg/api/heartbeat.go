package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	jww "github.com/spf13/jwalterweatherman"
)

// Send sends a bulk of heartbeats to the wakatime api.
func (c *Client) Send(heartbeats []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
	url := c.baseURL + "/v1/users/current/heartbeats.bulk"

	jww.DEBUG.Printf("sending %d heartbeat(s) to api at %q", len(heartbeats), url)

	data, err := json.Marshal(heartbeats)
	if err != nil {
		return nil, fmt.Errorf("failed to json encode body: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, Err(fmt.Sprintf("failed making request to %q: %s", url, err))
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, Err(fmt.Sprintf("failed reading response body from %q: %s", url, err))
	}

	switch resp.StatusCode {
	case http.StatusCreated, http.StatusAccepted:
		break
	case http.StatusUnauthorized:
		return nil, ErrAuth(fmt.Sprintf("authentication failed at %q", url))
	default:
		return nil, Err(fmt.Sprintf(
			"invalid response status from %q. got: %d, want: %d/%d. body: %q",
			url,
			resp.StatusCode,
			http.StatusCreated,
			http.StatusAccepted,
			string(body),
		))
	}

	results, err := ParseHeartbeatResponses(body)
	if err != nil {
		return nil, Err(fmt.Sprintf("failed parsing results from %q: %s", url, err))
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

	if result.Status == http.StatusBadRequest {
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
		Error  *string              `json:"error"`
		Errors *map[string][]string `json:"errors"`
	}

	// 1. try "error" key
	var resultError string

	err := json.Unmarshal(data, &responseBodyErr{Error: &resultError})
	if err != nil {
		return nil, fmt.Errorf("failed to parse json heartbeat error: %s", err)
	}

	if resultError != "" {
		errs = append(errs, resultError)
		return errs, nil
	}

	// 2. try "errors" key
	var resultErrors map[string][]string

	err = json.Unmarshal(data, &responseBodyErr{Errors: &resultErrors})
	if err != nil {
		return nil, fmt.Errorf("failed to parse json heartbeat errors: %s", err)
	}

	if resultErrors == nil {
		return nil, errors.New("failed to detect any errors despite invalid response status")
	}

	for field, messages := range resultErrors {
		errs = append(errs, fmt.Sprintf(
			"%s: %s",
			field,
			strings.Join(messages, " "),
		))
	}

	return errs, nil
}
