package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/summary"
)

const summaryDateFormat = "2006-01-02"

// SummaryConfig contains parameters for making a summaries request.
type SummaryConfig struct {
	Auth      BasicAuth
	UserAgent string
}

var (
	// Err represents a general api error.
	Err = errors.New("api error")
	// ErrAuth represents an authentication error.
	ErrAuth = errors.New("auth error")
)

// Client communicates with the wakatime api.
type Client struct {
	baseURL string
	client  *http.Client
}

// NewClient creates a new Client.
func NewClient(baseURL string, client *http.Client) *Client {
	return &Client{
		baseURL: baseURL,
		client:  client,
	}
}

// Summaries fetches summaries for the defined date range.
func (c *Client) Summaries(startDate, endDate time.Time, cfg SummaryConfig) ([]summary.Summary, error) {
	url := c.baseURL + "/api/v1/users/current/summaries"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	authHeaderValue, err := cfg.Auth.HeaderValue()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve auth header value: %s", err)
	}

	req.Header.Set("Authorization", authHeaderValue)

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", cfg.UserAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to %q: %w: %s", url, Err, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %q: %w: %s", url, Err, err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("authentication failed at %q: %w", url, ErrAuth)
	default:
		return nil, fmt.Errorf(
			"invalid response status from %q. got: %d, want: %d. body: %q: %w",
			url,
			resp.StatusCode,
			http.StatusOK,
			string(body),
			Err,
		)
	}

	summaries, err := parseSummariesResponse(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse results from %q: %w: %s", url, Err, err)
	}

	return summaries, nil
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
