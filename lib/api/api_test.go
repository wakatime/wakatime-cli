package api_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alanhamlett/wakatime-cli/lib/api"
	"github.com/alanhamlett/wakatime-cli/lib/heartbeat/subtypes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_SendHeartbeats(t *testing.T) {
	tests := []int{
		http.StatusCreated,
		http.StatusAccepted,
	}

	for _, code := range tests {
		t.Run(http.StatusText(code), func(t *testing.T) {
			url, router, close := setupTestServer()
			defer close()

			var numCalls int
			router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
				numCalls++

				// check headers
				assert.Equal(t, http.MethodPost, req.Method)
				assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
				assert.Equal(t, []string{"Basic c2VjcmV0"}, req.Header["Authorization"])
				assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])
				assert.Equal(t, []string{"wakatime/13.0.8"}, req.Header["User-Agent"])
				assert.Equal(t, []string{"MacBook"}, req.Header["X-Machine-Name"])

				// check body
				expectedBody, err := ioutil.ReadFile("testdata/api_heartbeats_request.json")
				require.NoError(t, err)

				body, err := ioutil.ReadAll(req.Body)
				require.NoError(t, err)

				assert.JSONEq(t, string(expectedBody), string(body))

				// write response
				f, err := os.Open("testdata/api_heartbeats_response.json")
				require.NoError(t, err)

				w.WriteHeader(code)
				_, err = io.Copy(w, f)
				require.NoError(t, err)
			})

			c := api.NewClient(url, http.DefaultClient)
			results, err := c.SendHeartbeats(testHeartbeats(), testConfig())
			require.NoError(t, err)

			// check via assert.Equal on complete slice here, to assert exact order of results,
			// which is assumed to exactly match the request order
			assert.Equal(t, []api.Result{
				{
					Status: http.StatusCreated,
					Heartbeat: api.Heartbeat{
						Branch:         api.String("heartbeat"),
						Category:       subtypes.CodingCategory,
						CursorPosition: api.Int(12),
						Dependencies:   []string{"dep1", "dep2"},
						Entity:         "/tmp/main.go",
						EntityType:     subtypes.FileType,
						IsWrite:        true,
						Language:       "golang",
						LineNumber:     api.Int(42),
						Lines:          api.Int(100),
						Project:        "wakatime-cli",
						Time:           1585598059,
						UserAgent:      "wakatime/13.0.6",
					},
				},
				{
					Status: http.StatusCreated,
					Heartbeat: api.Heartbeat{
						Branch:         nil,
						Category:       subtypes.DebuggingCategory,
						CursorPosition: nil,
						Dependencies:   nil,
						Entity:         "HIDDEN.py",
						EntityType:     subtypes.FileType,
						IsWrite:        false,
						Language:       "python",
						LineNumber:     nil,
						Lines:          nil,
						Project:        "wakatime",
						Time:           1585598060,
						UserAgent:      "wakatime/13.0.7",
					},
				},
			}, results)

			assert.Equal(t, 1, numCalls)
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

	c := api.NewClient(url, http.DefaultClient)
	_, err := c.SendHeartbeats(testHeartbeats(), testConfig())
	assert.IsType(t, api.Err{}, err)
}

func TestClient_SendHeartbeats_ErrAuth(t *testing.T) {
	url, router, close := setupTestServer()
	defer close()

	var numCalls int
	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusUnauthorized)
	})

	c := api.NewClient(url, http.DefaultClient)
	_, err := c.SendHeartbeats(testHeartbeats(), testConfig())
	assert.IsType(t, api.ErrAuth{}, err)
}

func testConfig() api.Config {
	return api.Config{
		APIKey:    "secret",
		HostName:  "MacBook",
		UserAgent: "wakatime/13.0.8",
	}
}

func testHeartbeats() []api.Heartbeat {
	return []api.Heartbeat{
		{
			Branch:         api.String("heartbeat"),
			Category:       subtypes.CodingCategory,
			CursorPosition: api.Int(12),
			Dependencies:   []string{"dep1", "dep2"},
			Entity:         "/tmp/main.go",
			EntityType:     subtypes.FileType,
			IsWrite:        true,
			Language:       "golang",
			LineNumber:     api.Int(42),
			Lines:          api.Int(100),
			Project:        "wakatime-cli",
			Time:           1585598059,
			UserAgent:      "wakatime/13.0.6",
		},
		{
			Branch:         nil,
			Category:       subtypes.DebuggingCategory,
			CursorPosition: nil,
			Dependencies:   nil,
			Entity:         "HIDDEN.py",
			EntityType:     subtypes.FileType,
			IsWrite:        false,
			Language:       "python",
			LineNumber:     nil,
			Lines:          nil,
			Project:        "wakatime",
			Time:           1585598060,
			UserAgent:      "wakatime/13.0.7",
		},
	}
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}
