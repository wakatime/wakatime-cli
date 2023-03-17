package fileexperts_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/cmd"
	"github.com/wakatime/wakatime-cli/cmd/fileexperts"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileExperts(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	projectFolder, err := filepath.Abs("../..")
	require.NoError(t, err)

	subfolders := project.CountSlashesInProjectFolder(projectFolder)

	router.HandleFunc("/users/current/file_experts", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check request
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.True(t, strings.HasSuffix(req.Header["User-Agent"][0], plugin), fmt.Sprintf(
			"%q should have suffix %q",
			req.Header["User-Agent"][0],
			plugin,
		))

		expectedBody, err := os.ReadFile("testdata/api_file_experts_request_template.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var entity struct {
			Entity string `json:"entity"`
		}

		err = json.Unmarshal(body, &entity)
		require.NoError(t, err)

		expectedBodyStr := fmt.Sprintf(string(expectedBody), entity.Entity, subfolders)

		assert.True(t, strings.HasSuffix(entity.Entity, "testdata/main.go"))
		assert.JSONEq(t, expectedBodyStr, string(body))

		// send response
		w.WriteHeader(http.StatusOK)

		f, err := os.Open("testdata/api_file_experts_response.json")
		require.NoError(t, err)
		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)
	v.Set("plugin", plugin)
	v.Set("project", "wakatime-cli")
	v.Set("entity", "testdata/main.go")
	v.Set("file-experts", true)

	output, err := fileexperts.FileExperts(v)
	require.NoError(t, err)

	assert.Equal(t, "You: 40 mins | Karl: 21 mins", output)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestFileExperts_NonExistingEntity(t *testing.T) {
	tmpDir := t.TempDir()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer logFile.Close()

	v := viper.New()
	v.Set("api-url", "https://example.org")
	v.Set("entity", "nonexisting")
	v.Set("file-experts", true)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("log-file", logFile.Name())
	v.Set("verbose", true)

	cmd.SetupLogging(v)

	defer func() {
		if file, ok := log.Output().(*os.File); ok {
			_ = file.Sync()
			file.Close()
		} else if handler, ok := log.Output().(io.Closer); ok {
			handler.Close()
		}
	}()

	_, err = fileexperts.FileExperts(v)
	require.NoError(t, err)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "skipping because of non-existing file")
}

func TestFileExperts_ErrApi(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/file_experts", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusInternalServerError)
	})

	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)
	v.Set("entity", "testdata/main.go")
	v.Set("file-experts", true)

	_, err := fileexperts.FileExperts(v)
	require.Error(t, err)

	var errapi api.Err

	assert.ErrorAs(t, err, &errapi)

	expectedMsg := fmt.Sprintf(
		`invalid response status from "%s/users/current/file_experts". got: 500, want: 200. body: ""`,
		testServerURL,
	)

	assert.EqualError(t, err, expectedMsg)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestFileExperts_ErrAuth(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/file_experts", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusUnauthorized)
	})

	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)
	v.Set("entity", "testdata/main.go")
	v.Set("file-experts", true)

	_, err := fileexperts.FileExperts(v)
	require.Error(t, err)

	var errauth api.ErrAuth

	assert.ErrorAs(t, err, &errauth)

	expectedMsg := fmt.Sprintf(
		`authentication failed at "%s/users/current/file_experts". body: ""`,
		testServerURL,
	)
	assert.EqualError(t, err, expectedMsg)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestFileExperts_ErrBadRequest(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/file_experts", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusBadRequest)
	})

	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)
	v.Set("entity", "testdata/main.go")
	v.Set("file-experts", true)

	_, err := fileexperts.FileExperts(v)
	require.Error(t, err)

	var errbadRequest api.ErrBadRequest

	assert.ErrorAs(t, err, &errbadRequest)

	expectedMsg := fmt.Sprintf(
		`bad request at "%s/users/current/file_experts"`,
		testServerURL,
	)

	assert.EqualError(t, err, expectedMsg)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}
