package api_test

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHeartbeatResponses(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/api_heartbeats_response.json")
	require.NoError(t, err)

	results, err := api.ParseHeartbeatResponses(data)
	require.NoError(t, err)

	// check via assert.Equal on complete slice here, to assert exact order of results,
	// which is assumed to exactly match the request order
	assert.Equal(t, results, []heartbeat.Result{
		{
			Status: http.StatusCreated,
			Heartbeat: heartbeat.Heartbeat{
				Branch:         heartbeat.String("heartbeat"),
				Category:       heartbeat.CodingCategory,
				CursorPosition: heartbeat.Int(12),
				Dependencies:   []string{"dep1", "dep2"},
				Entity:         "/tmp/main.go",
				EntityType:     heartbeat.FileType,
				IsWrite:        heartbeat.Bool(true),
				Language:       heartbeat.String("golang"),
				LineNumber:     heartbeat.Int(42),
				Lines:          heartbeat.Int(100),
				Project:        heartbeat.String("wakatime-cli"),
				Time:           1585598059,
				UserAgent:      "wakatime/13.0.6",
			},
		},
		{
			Status: http.StatusCreated,
			Heartbeat: heartbeat.Heartbeat{
				Branch:         nil,
				Category:       heartbeat.DebuggingCategory,
				CursorPosition: nil,
				Dependencies:   nil,
				Entity:         "HIDDEN.py",
				EntityType:     heartbeat.FileType,
				IsWrite:        nil,
				Language:       nil,
				LineNumber:     nil,
				Lines:          nil,
				Project:        nil,
				Time:           1585598060,
				UserAgent:      "wakatime/13.0.7",
			},
		},
	})
}

func TestParseHeartbeatResponsesErr(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/api_heartbeats_response_error.json")
	require.NoError(t, err)

	results, err := api.ParseHeartbeatResponses(data)
	require.NoError(t, err)

	// asserting here the exact order of results, which is assumed to exactly match the request order
	assert.Len(t, results, 2)

	assert.Equal(t, 400, results[0].Status)
	assert.Len(t, results[0].Errors, 2)
	assert.Contains(t, results[0].Errors, "lineno: Number must be between 1 and 2147483647.")
	assert.Contains(t, results[0].Errors, "time: This field is required.")

	assert.Equal(t, heartbeat.Result{
		Errors: []string{"Can not log time before user was created."},
		Status: 400,
	}, results[1])
}
