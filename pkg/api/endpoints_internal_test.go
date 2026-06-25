package api

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndpointReadAndParseErrors(t *testing.T) {
	t.Run("today read error", func(t *testing.T) {
		c := clientWithResponse(http.StatusOK, errorReadCloser{})

		_, err := c.Today(t.Context())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read response body")
	})

	t.Run("today parse error", func(t *testing.T) {
		c := clientWithResponse(http.StatusOK, io.NopCloser(strings.NewReader(`{`)))

		_, err := c.Today(t.Context())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse results")
	})

	t.Run("goal read error", func(t *testing.T) {
		c := clientWithResponse(http.StatusOK, errorReadCloser{})

		_, err := c.Goal(t.Context(), "goal")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read response body")
	})

	t.Run("goal parse error", func(t *testing.T) {
		c := clientWithResponse(http.StatusOK, io.NopCloser(strings.NewReader(`{`)))

		_, err := c.Goal(t.Context(), "goal")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse results")
	})

	t.Run("file experts read error", func(t *testing.T) {
		c := clientWithResponse(http.StatusOK, errorReadCloser{})

		_, err := c.FileExperts(t.Context(), []heartbeat.Heartbeat{{APIKey: "key", Entity: "main.go"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed reading response body")
	})

	t.Run("file experts parse error", func(t *testing.T) {
		c := clientWithResponse(http.StatusOK, io.NopCloser(strings.NewReader(`{`)))

		_, err := c.FileExperts(t.Context(), []heartbeat.Heartbeat{{APIKey: "key", Entity: "main.go"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse results")
	})
}

func TestParseEndpointResponsesInvalidJSON(t *testing.T) {
	_, err := ParseStatusBarResponse([]byte(`{`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse json response body")

	_, err = ParseGoalResponse([]byte(`{`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse json response body")

	_, err = ParseFileExpertsResponse([]byte(`{`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse json response body")
}

func clientWithResponse(status int, body io.ReadCloser) *Client {
	c := NewClient(BaseURL)
	c.doFunc = func(*Client, *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Body:       body,
		}, nil
	}

	return c
}
