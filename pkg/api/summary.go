package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/summary"
)

const dateFormat = "2006-01-02"

// Summaries fetches summaries for the defined date range.
//
// ErrRequest is returned upon request failure with no received response from api.
// ErrAuth is returned upon receiving a 401 Unauthorized api response.
// Err is returned on any other api response related error.
func (c *Client) Summaries(startDate, endDate time.Time) ([]summary.Summary, error) {
	url := c.baseURL + "/users/current/summaries"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	q := req.URL.Query()
	q.Add("start", startDate.Format(dateFormat))
	q.Add("end", endDate.Format(dateFormat))
	req.URL.RawQuery = q.Encode()

	resp, err := c.Do(req)
	if err != nil {
		return nil, ErrRequest(fmt.Sprintf("failed to make request to %q: %s", url, err))
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, Err(fmt.Sprintf("failed to read response body from %q: %s", url, err))
	}

	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusUnauthorized:
		return nil, ErrAuth(fmt.Sprintf("authentication failed at %q. body: %q", url, string(body)))
	default:
		return nil, Err(fmt.Sprintf(
			"invalid response status from %q. got: %d, want: %d. body: %q",
			url,
			resp.StatusCode,
			http.StatusOK,
			string(body),
		))
	}

	summaries, err := ParseSummariesResponse(body)
	if err != nil {
		return nil, Err(fmt.Sprintf("failed to parse results from %q: %s", url, err))
	}

	return summaries, nil
}

// ParseSummariesResponse parses the wakatime api response into summary.Summary.
func ParseSummariesResponse(data []byte) ([]summary.Summary, error) {
	var body struct {
		Data []struct {
			Categories []struct {
				Name string `json:"name"`
				Text string `json:"text"`
			} `json:"categories"`
			GrandTotal struct {
				Text string `json:"text"`
			} `json:"grand_total"`
			Range struct {
				Date string `json:"date"`
			} `json:"range"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &body); err != nil {
		return nil, fmt.Errorf("failed to parse json response body: %s. body: %q", err, data)
	}

	var summaries []summary.Summary

	for _, sum := range body.Data {
		date, err := time.Parse(dateFormat, sum.Range.Date)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date from string %q: %s", sum.Range.Date, err)
		}

		parsed := summary.Summary{
			Date:  date,
			Total: sum.GrandTotal.Text,
		}

		if len(sum.Categories) > 0 {
			for _, category := range sum.Categories {
				parsed.ByCategory = append(parsed.ByCategory, summary.Category{
					Category: category.Name,
					Total:    category.Text,
				})
			}
		}

		summaries = append(summaries, parsed)
	}

	return summaries, nil
}
