package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/wakatime/wakatime-cli/pkg/summary"
)

// Today fetches code stats for Today.
//
// ErrRequest is returned upon request failure with no received response from api.
// ErrAuth is returned upon receiving a 401 Unauthorized api response.
// Err is returned on any other api response related error.
func (c *Client) Today() (*summary.Summary, error) {
	url := c.baseURL + "/users/current/statusbar/today"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, Err{fmt.Errorf("failed to create request: %s", err)}
	}

	q := req.URL.Query()
	req.URL.RawQuery = q.Encode()

	resp, err := c.Do(req)
	if err != nil {
		return nil, Err{fmt.Errorf("failed to make request to %q: %s", url, err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, Err{fmt.Errorf("failed to read response body from %q: %s", url, err)}
	}

	switch resp.StatusCode {
	case http.StatusOK:
		break
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

	summary, err := ParseSummaryResponse(body)
	if err != nil {
		return nil, Err{Err: fmt.Errorf("failed to parse results from %q: %s", url, err)}
	}

	return summary, nil
}

// ParseSummaryResponse parses the wakatime api response into summary.Summary.
func ParseSummaryResponse(data []byte) (*summary.Summary, error) {
	var body struct {
		Data struct {
			Categories []struct {
				Name string `json:"name"`
				Text string `json:"text"`
			} `json:"categories"`
			GrandTotal struct {
				Text string `json:"text"`
			} `json:"grand_total"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &body); err != nil {
		return nil, fmt.Errorf("failed to parse json response body: %s. body: %q", err, data)
	}

	parsed := summary.Summary{
		Total: body.Data.GrandTotal.Text,
	}

	if len(body.Data.Categories) > 0 {
		for _, category := range body.Data.Categories {
			parsed.ByCategory = append(parsed.ByCategory, summary.Category{
				Category: category.Name,
				Total:    category.Text,
			})
		}
	}

	return &parsed, nil
}
