package fileexperts_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wakatime/wakatime-cli/cmd/fileexperts"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/log/setup"
	"github.com/wakatime/wakatime-cli/pkg/project"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	router.HandleFunc("/users/current/file_experts", func(w http.ResponseWriter, _ *http.Request) {
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
	v.Set("entity", "testdata/main.go")
	v.Set("projectmap..*", "wakatime-cli")

	var (
		code int
		err  error
	)

	output := captureStdout(t, func() {
		code, err = fileexperts.Run(t.Context(), v)
	})

	require.NoError(t, err)
	assert.Equal(t, exitcode.Success, code)
	assert.Equal(t, "You: 40 mins | Karl: 21 mins\n", output)
}

func TestRunErr(t *testing.T) {
	v := viper.New()
	v.Set("entity", "testdata/main.go")

	code, err := fileexperts.Run(t.Context(), v)

	require.Error(t, err)
	assert.Equal(t, exitcode.ErrGeneric, code)
	assert.Contains(t, err.Error(), "file experts fetch failed")
}

func TestFileExperts(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

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

		expectedBodyStr := fmt.Sprintf(
			string(expectedBody),
			entity.Entity,
			project.CountSlashesInProjectFolder(filepath.Dir(entity.Entity)),
		)

		assert.True(t, strings.HasSuffix(entity.Entity, "testdata/main.go"))
		assertJSONEqIgnoringProjectRootCount(t, expectedBodyStr, string(body))

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
	v.Set("projectmap..*", "wakatime-cli")
	v.Set("entity", "testdata/main.go")

	output, err := fileexperts.FileExperts(t.Context(), v)
	require.NoError(t, err)

	assert.Equal(t, "You: 40 mins | Karl: 21 mins", output)

	assert.Equal(t, 1, numCalls)
}

func TestFileExperts_NonExistingEntity(t *testing.T) {
	ctx := t.Context()

	logFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer logFile.Close()

	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", "https://example.org")
	v.Set("entity", "nonexisting")
	v.Set("log-file", logFile.Name())
	v.Set("verbose", true)

	logger, err := setup.Logging(ctx, v)
	require.NoError(t, err)

	defer logger.Flush()

	ctx = log.ToContext(ctx, logger)

	_, err = fileexperts.FileExperts(ctx, v)
	require.NoError(t, err)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "skipping because of non-existing file")
}

func TestFileExperts_ErrApi(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/file_experts", func(w http.ResponseWriter, _ *http.Request) {
		numCalls++

		w.WriteHeader(http.StatusInternalServerError)
	})

	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)
	v.Set("entity", "testdata/main.go")
	v.Set("projectmap..*", "wakatime-cli")

	_, err := fileexperts.FileExperts(t.Context(), v)
	require.Error(t, err)

	var errapi api.Err

	assert.ErrorAs(t, err, &errapi)

	expectedMsg := fmt.Sprintf(
		`invalid response status from "%s/users/current/file_experts". got: 500, want: 200. body: ""`,
		testServerURL,
	)

	assert.EqualError(t, err, expectedMsg)

	assert.Equal(t, 1, numCalls)
}

func TestFileExperts_ErrAuth(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/file_experts", func(w http.ResponseWriter, _ *http.Request) {
		numCalls++

		w.WriteHeader(http.StatusUnauthorized)
	})

	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)
	v.Set("entity", "testdata/main.go")
	v.Set("projectmap..*", "wakatime-cli")

	_, err := fileexperts.FileExperts(t.Context(), v)
	require.Error(t, err)

	var errauth api.ErrAuth

	assert.ErrorAs(t, err, &errauth)

	expectedMsg := fmt.Sprintf(
		`authentication failed at "%s/users/current/file_experts". body: ""`,
		testServerURL,
	)
	assert.EqualError(t, err, expectedMsg)

	assert.Equal(t, 1, numCalls)
}

func TestFileExperts_ErrBadRequest(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/file_experts", func(w http.ResponseWriter, _ *http.Request) {
		numCalls++

		w.WriteHeader(http.StatusBadRequest)
	})

	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)
	v.Set("entity", "testdata/main.go")
	v.Set("projectmap..*", "wakatime-cli")

	_, err := fileexperts.FileExperts(t.Context(), v)
	require.Error(t, err)

	var errbadRequest api.ErrBadRequest

	assert.ErrorAs(t, err, &errbadRequest)

	expectedMsg := fmt.Sprintf(
		`bad request at "%s/users/current/file_experts"`,
		testServerURL,
	)

	assert.EqualError(t, err, expectedMsg)

	assert.Equal(t, 1, numCalls)
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}

func assertJSONEqIgnoringProjectRootCount(t *testing.T, expected, actual string) {
	t.Helper()

	var expectedJSON any
	require.NoError(t, json.Unmarshal([]byte(expected), &expectedJSON))

	var actualJSON any
	require.NoError(t, json.Unmarshal([]byte(actual), &actualJSON))

	removeProjectRootCount(expectedJSON)
	removeProjectRootCount(actualJSON)

	assert.Equal(t, expectedJSON, actualJSON)
}

func removeProjectRootCount(v any) {
	switch vv := v.(type) {
	case map[string]any:
		delete(vv, "project_root_count")

		for _, child := range vv {
			removeProjectRootCount(child)
		}
	case []any:
		for _, child := range vv {
			removeProjectRootCount(child)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	stdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = w

	defer func() { os.Stdout = stdout }()

	outC := make(chan string, 1)
	errC := make(chan error, 1)

	go func() {
		var buf bytes.Buffer

		_, err := io.Copy(&buf, r)
		errC <- err

		outC <- buf.String()
	}()

	fn()

	require.NoError(t, w.Close())

	output := <-outC

	require.NoError(t, <-errC)
	require.NoError(t, r.Close())

	return output
}
