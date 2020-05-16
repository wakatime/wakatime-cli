package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/summary"
)

const summaryDateFormat = "2006-01-02"

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

// Summaries fetches summaries for the defined date range.
func (c *Client) Summaries(startDate, endDate time.Time) ([]summary.Summary, error) {
	url := c.baseURL + "/api/v1/users/current/summaries"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, Err(fmt.Sprintf("failed to make request to %q: %s", url, err))
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
		return nil, ErrAuth(fmt.Sprintf("authentication failed at %q", url))
	default:
		return nil, Err(fmt.Sprintf(
			"invalid response status from %q. got: %d, want: %d. body: %q",
			url,
			resp.StatusCode,
			http.StatusOK,
			string(body),
		))
	}

	summaries, err := parseSummariesResponse(body)
	if err != nil {
		return nil, Err(fmt.Sprintf("failed to parse results from %q: %s", url, err))
	}

	return summaries, nil
}

// do wraps c.client.Do() and sets default headers and headers, which are set
// via Option.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Accept", "application/json")

	if c.authHeader != "" {
		req.Header.Set("Authorization", c.authHeader)
	}

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	return c.client.Do(req)
}

// parseSummariesResponse parses the wakatime api response into summary.Summary.
func parseSummariesResponse(data []byte) ([]summary.Summary, error) {
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
		return nil, fmt.Errorf("failed to json unmarshal response body: %s. body: %q", err, data)
	}

	var summaries []summary.Summary

	for _, sum := range body.Data {
		date, err := time.Parse(summaryDateFormat, sum.Range.Date)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date from string %q: %s", sum.Range.Date, err)
		}

		if len(sum.Categories) > 0 {
			for _, category := range sum.Categories {
				summaries = append(summaries, summary.Summary{
					Category:   category.Name,
					GrandTotal: category.Text,
					Date:       date,
				})
			}

			continue
		}

		summaries = append(summaries, summary.Summary{
			GrandTotal: sum.GrandTotal.Text,
			Date:       date,
		})
	}

	return summaries, nil
}
