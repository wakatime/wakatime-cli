package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type errorReadCloser struct{}

func (errorReadCloser) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func (errorReadCloser) Close() error {
	return nil
}

func TestParseHeartbeatResponsesBranches(t *testing.T) {
	_, err := ParseHeartbeatResponses(t.Context(), []byte(`{`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse json response body")

	_, err = ParseHeartbeatResponses(t.Context(), []byte(`{"responses":[[{}, "bad"]]}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed parsing result #0")

	_, err = parseHeartbeatResponse(t.Context(), []json.RawMessage{jsonRaw(`{}`), jsonRaw(`"bad"`)})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse json status")

	_, err = parseHeartbeatResponse(t.Context(), []json.RawMessage{jsonRaw(`{`), jsonRaw(`201`)})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse json heartbeat")

	_, err = parseHeartbeatResponse(t.Context(), []json.RawMessage{jsonRaw(`{`), jsonRaw(`500`)})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse result errors")

	result, err := parseHeartbeatResponse(t.Context(), []json.RawMessage{
		jsonRaw(`{"errors":{"dependencies":["skip"],"entity":["bad","missing"]}}`),
		jsonRaw(`400`),
	})
	require.NoError(t, err)
	assert.Equal(t, 400, result.Status)
	assert.Equal(t, []string{"entity: bad missing"}, result.Errors)
}

func TestClientSendHeartbeatsInternalBranches(t *testing.T) {
	t.Run("response body read error", func(t *testing.T) {
		c := NewClient(BaseURL)
		c.doFunc = func(*Client, *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusCreated,
				Body:       errorReadCloser{},
			}, nil
		}

		_, err := c.SendHeartbeats(t.Context(), []heartbeat.Heartbeat{{APIKey: "key"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed reading response body")
	})

	t.Run("parse response error", func(t *testing.T) {
		c := NewClient(BaseURL)
		c.doFunc = func(*Client, *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(strings.NewReader(`{`)),
			}, nil
		}

		_, err := c.SendHeartbeats(t.Context(), []heartbeat.Heartbeat{{APIKey: "key"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed parsing results")
	})

	t.Run("extra results are not assigned heartbeats", func(t *testing.T) {
		c := NewClient(BaseURL)
		c.doFunc = func(*Client, *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusCreated,
				Body: io.NopCloser(strings.NewReader(`{"responses":[` +
					`[{"data":{"id":"first"}},201],` +
					`[{"data":{"id":"extra"}},201]` +
					`]}`)),
			}, nil
		}

		input := heartbeat.Heartbeat{APIKey: "key", Entity: "main.go"}
		results, err := c.SendHeartbeats(context.Background(), []heartbeat.Heartbeat{input})
		require.NoError(t, err)
		require.Len(t, results, 2)
		assert.Equal(t, input, results[0].Heartbeat)
		assert.Empty(t, results[1].Heartbeat)
	})
}

func jsonRaw(value string) json.RawMessage {
	return json.RawMessage(value)
}
