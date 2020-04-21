package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type Result struct {
	Errors    []string
	Status    int
	Heartbeat Heartbeat
}

func parseResults(data []byte) ([]Result, error) {
	var responseBody struct {
		Responses [][]json.RawMessage `json:"responses"`
	}

	err := json.Unmarshal(data, &responseBody)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal response body: %s", err)
	}

	var results []Result
	for _, r := range responseBody.Responses {
		result, err := parseResult(r)
		if err != nil {
			return nil, fmt.Errorf("failed parsing result: %s", err)
		}

		results = append(results, result)
	}

	return results, nil
}

// response is used to parse heartbeat responses from the send heartbeats bulk response body
// response body examples:
//   - ./testdata/api_heartbeats_response.json
//   - ./testdata/api_heartbeats_response_error.json
type response struct {
	Error  *string              `json:"error"`
	Errors *map[string][]string `json:"errors"`
	Data   *Heartbeat           `json:"data"`
}

func parseResult(data []json.RawMessage) (Result, error) {
	var result Result
	err := json.Unmarshal(data[1], &result.Status)
	if err != nil {
		return Result{}, fmt.Errorf("failed json unmarshal status: %s", err)
	}

	if result.Status == http.StatusBadRequest {
		resultErrors, err := parseResultErrors(data[0])
		if err != nil {
			return Result{}, fmt.Errorf("failed parsing result errors: %s", err)
		}

		result.Errors = resultErrors
		return result, nil
	}

	err = json.Unmarshal(data[0], &response{Data: &result.Heartbeat})
	if err != nil {
		return Result{}, fmt.Errorf("failed json unmarshal heartbeat: %s", err)
	}

	return result, nil
}

func parseResultErrors(data json.RawMessage) ([]string, error) {
	var errs []string

	// 1. try "error" key
	var resultError string
	err := json.Unmarshal(data, &response{Error: &resultError})
	if err != nil {
		return nil, fmt.Errorf("failed json unmarshal heartbeat error: %s", err)
	}

	if resultError != "" {
		errs = append(errs, resultError)
		return errs, nil
	}

	// 2. try "errors" key
	var resultErrors map[string][]string
	err = json.Unmarshal(data, &response{Errors: &resultErrors})
	if err != nil {
		return nil, fmt.Errorf("failed json unmarshal heartbeat errors: %s", err)
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
