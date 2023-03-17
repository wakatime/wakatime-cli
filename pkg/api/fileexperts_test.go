package api_test

import (
	"errors"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/fileexperts"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_FileExperts(t *testing.T) {
	tests := []int{
		http.StatusOK,
		http.StatusAccepted,
	}

	for _, status := range tests {
		t.Run(http.StatusText(status), func(t *testing.T) {
			url, router, close := setupTestServer()
			defer close()

			var numCalls int

			router.HandleFunc("/users/current/file_experts", func(w http.ResponseWriter, req *http.Request) {
				numCalls++

				// check headers
				assert.Equal(t, http.MethodPost, req.Method)
				assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
				assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])

				// check body
				expectedBody, err := os.ReadFile("testdata/api_file_experts_request.json")
				require.NoError(t, err)

				body, err := io.ReadAll(req.Body)
				require.NoError(t, err)

				assert.JSONEq(t, string(expectedBody), string(body))

				// write response
				f, err := os.Open("testdata/api_file_experts_response.json")
				require.NoError(t, err)

				w.WriteHeader(status)
				_, err = io.Copy(w, f)
				require.NoError(t, err)
			})

			c := api.NewClient(url)
			results, err := c.FileExperts([]heartbeat.Heartbeat{
				{
					APIKey:           "00000000-0000-4000-8000-000000000000",
					Entity:           "/tmp/main.go",
					Project:          heartbeat.PointerTo("wakatime-cli"),
					ProjectRootCount: heartbeat.PointerTo(6),
				},
			})
			require.NoError(t, err)

			assert.Len(t, results, 1)

			assert.Equal(t, &fileexperts.FileExperts{
				Data: []fileexperts.Data{
					{
						Total: fileexperts.Total{
							Decimal:      "0.67",
							Digital:      "0:40",
							Text:         "40 mins",
							TotalSeconds: 2409,
						},
						User: fileexperts.User{
							ID:            "4b023c6f-f2f8-4212-94ee-48eb5f8f5c94",
							IsCurrentUser: true,
							LongName:      "John Doe",
							Name:          "John",
						},
					},
					{
						Total: fileexperts.Total{
							Decimal:      "0.35",
							Digital:      "0:21",
							Text:         "21 mins",
							TotalSeconds: 1301,
						},
						User: fileexperts.User{
							ID:       "f550f8d6-6e83-454f-be58-1d4a0b1ec81b",
							LongName: "Karl Marx",
							Name:     "Karl",
						},
					},
					{
						Total: fileexperts.Total{
							Decimal:      "0.00",
							Digital:      "0:00",
							Text:         "0 secs",
							TotalSeconds: 0,
						},
						User: fileexperts.User{
							ID:       "f14f298d-86b0-4eb8-a23d-4fda2596f035",
							LongName: "Nick Fury",
							Name:     "Nick",
						},
					},
				},
			}, results[0].FileExpert)

			assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
		})
	}
}

func TestClient_FileExperts_Err(t *testing.T) {
	url, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/users/current/file_experts", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusInternalServerError)
	})

	c := api.NewClient(url)
	_, err := c.FileExperts([]heartbeat.Heartbeat{
		{
			APIKey:           "00000000-0000-4000-8000-000000000000",
			Entity:           "/tmp/main.go",
			Project:          heartbeat.PointerTo("wakatime-cli"),
			ProjectRootCount: heartbeat.PointerTo(6),
		},
	})

	var errapi api.Err

	assert.True(t, errors.As(err, &errapi))

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_FileExperts_ErrAuth(t *testing.T) {
	url, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/users/current/file_experts", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusUnauthorized)
	})

	c := api.NewClient(url)
	_, err := c.FileExperts([]heartbeat.Heartbeat{
		{
			APIKey:           "00000000-0000-4000-8000-000000000000",
			Entity:           "/tmp/main.go",
			Project:          heartbeat.PointerTo("wakatime-cli"),
			ProjectRootCount: heartbeat.PointerTo(6),
		},
	})

	var errauth api.ErrAuth

	assert.ErrorAs(t, err, &errauth)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_FileExperts_ErrBadRequest(t *testing.T) {
	url, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/users/current/file_experts", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusBadRequest)
	})

	c := api.NewClient(url)
	_, err := c.FileExperts([]heartbeat.Heartbeat{
		{
			APIKey:           "00000000-0000-4000-8000-000000000000",
			Entity:           "/tmp/main.go",
			Project:          heartbeat.PointerTo("wakatime-cli"),
			ProjectRootCount: heartbeat.PointerTo(6),
		},
	})

	var errbadRequest api.ErrBadRequest

	assert.True(t, errors.As(err, &errbadRequest))

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_FileExperts_InvalidUrl(t *testing.T) {
	c := api.NewClient("invalid-url")
	_, err := c.FileExperts([]heartbeat.Heartbeat{
		{
			APIKey:           "00000000-0000-4000-8000-000000000000",
			Entity:           "/tmp/main.go",
			Project:          heartbeat.PointerTo("wakatime-cli"),
			ProjectRootCount: heartbeat.PointerTo(6),
		},
	})

	var apierr api.Err

	assert.True(t, errors.As(err, &apierr))
}
