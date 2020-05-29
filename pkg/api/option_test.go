package api_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOption_WithHostName(t *testing.T) {
	url, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, []string{"my-computer"}, req.Header["X-Machine-Name"])

		numCalls++
	})

	opts := []api.Option{
		api.WithHostName("my-computer"),
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	c := api.NewClient("", http.DefaultClient, opts...)
	resp, err := c.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}
