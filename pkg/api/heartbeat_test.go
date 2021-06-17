package api_test

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_SendHeartbeats(t *testing.T) {
	tests := []int{
		http.StatusCreated,
		http.StatusAccepted,
	}

	for _, status := range tests {
		t.Run(http.StatusText(status), func(t *testing.T) {
			url, router, close := setupTestServer()
			defer close()

			var numCalls int

			router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
				numCalls++

				// check headers
				assert.Equal(t, http.MethodPost, req.Method)
				assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
				assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])

				// check body
				expectedBody, err := ioutil.ReadFile("testdata/api_heartbeats_request.json")
				require.NoError(t, err)

				body, err := ioutil.ReadAll(req.Body)
				require.NoError(t, err)

				assert.JSONEq(t, string(expectedBody), string(body))

				// write response
				f, err := os.Open("testdata/api_heartbeats_response.json")
				require.NoError(t, err)

				w.WriteHeader(status)
				_, err = io.Copy(w, f)
				require.NoError(t, err)
			})

			c := api.NewClient(url)
			results, err := c.SendHeartbeats(testHeartbeats())
			require.NoError(t, err)

			// check via assert.Equal on complete slice here, to assert exact order of results,
			// which is assumed to exactly match the request order
			assert.Equal(t, []heartbeat.Result{
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
						Language:       heartbeat.String("Go"),
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
						LineNumber:     nil,
						Lines:          nil,
						Project:        nil,
						Time:           1585598060,
						UserAgent:      "wakatime/13.0.7",
					},
				},
			}, results)

			assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
		})
	}
}

func TestClient_SendHeartbeats_Err(t *testing.T) {
	url, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusInternalServerError)
	})

	c := api.NewClient(url)
	_, err := c.SendHeartbeats(testHeartbeats())

	var errapi api.Err

	assert.True(t, errors.As(err, &errapi))

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_SendHeartbeats_ErrAuth(t *testing.T) {
	url, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusUnauthorized)
	})

	c := api.NewClient(url)
	_, err := c.SendHeartbeats(testHeartbeats())

	var errauth api.ErrAuth

	assert.True(t, errors.As(err, &errauth))

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_SendHeartbeats_ErrRequest(t *testing.T) {
	c := api.NewClient("invalid-url")
	_, err := c.SendHeartbeats(testHeartbeats())

	var errreq api.ErrRequest

	assert.True(t, errors.As(err, &errreq))
}

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
				Language:       heartbeat.String("Go"),
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
				LineNumber:     nil,
				Lines:          nil,
				Project:        nil,
				Time:           1585598060,
				UserAgent:      "wakatime/13.0.7",
			},
		},
	})
}

func TestParseHeartbeatResponses_Error(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/api_heartbeats_response_error.json")
	require.NoError(t, err)

	results, err := api.ParseHeartbeatResponses(data)
	require.NoError(t, err)

	// asserting here the exact order of results, which is assumed to exactly match the request order
	assert.Len(t, results, 4)

	// valid responses
	assert.Equal(t, 201, results[0].Status)
	assert.Equal(t, 201, results[1].Status)

	// error responses
	assert.Equal(t, 429, results[2].Status)
	assert.Equal(t, results[2].Errors, []string{"Too many heartbeats"})
	assert.Equal(t, 429, results[3].Status)
	assert.Equal(t, results[3].Errors, []string{"Too many heartbeats"})
}

func TestParseHeartbeatResponses_Errors(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/api_heartbeats_response_errors.json")
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

func testHeartbeats() []heartbeat.Heartbeat {
	return []heartbeat.Heartbeat{
		{
			Branch:         heartbeat.String("heartbeat"),
			Category:       heartbeat.CodingCategory,
			CursorPosition: heartbeat.Int(12),
			Dependencies:   []string{"dep1", "dep2"},
			Entity:         "/tmp/main.go",
			EntityType:     heartbeat.FileType,
			IsWrite:        heartbeat.Bool(true),
			Language:       heartbeat.String("Go"),
			LineNumber:     heartbeat.Int(42),
			Lines:          heartbeat.Int(100),
			Project:        heartbeat.String("wakatime-cli"),
			Time:           1585598059,
			UserAgent:      "wakatime/13.0.6",
		},
		{
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
	}
}
