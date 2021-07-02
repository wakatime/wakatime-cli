package heartbeat_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/cmd/legacy"
	cmd "github.com/wakatime/wakatime-cli/cmd/legacy/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendHeartbeats(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
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

		expectedBody, err := ioutil.ReadFile("testdata/api_heartbeats_request_template.json")
		require.NoError(t, err)

		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)

		var entity struct {
			Entity string `json:"entity"`
		}

		err = json.Unmarshal(body, &[]interface{}{&entity})
		require.NoError(t, err)

		expectedBodyStr := fmt.Sprintf(string(expectedBody), entity.Entity, heartbeat.UserAgent(plugin))

		assert.True(t, strings.HasSuffix(entity.Entity, "testdata/main.go"))
		assert.JSONEq(t, expectedBodyStr, string(body))

		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)
		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)

		numCalls++
	})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", testServerURL)
	v.Set("category", "debugging")
	v.Set("cursorpos", 42)
	v.Set("entity", "testdata/main.go")
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("language", "Go")
	v.Set("alternate-language", "Golang")
	v.Set("hide-branch-names", "true")
	v.Set("project", "wakatime-cli")
	v.Set("lineno", 13)
	v.Set("local-file", "testdata/localfile.go")
	v.Set("plugin", plugin)
	v.Set("time", 1585598059.1)
	v.Set("timeout", 5)
	v.Set("write", true)

	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(f.Name())

	err = cmd.SendHeartbeats(v, f.Name())
	require.NoError(t, err)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_WithFiltering_Exclude(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(500)

		numCalls++
	})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", testServerURL)
	v.Set("category", "debugging")
	v.Set("entity", "/tmp/main.go")
	v.Set("exclude", ".*")
	v.Set("entity-type", "app")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("plugin", "plugin")
	v.Set("time", 1585598059.1)
	v.Set("timeout", 5)
	v.Set("write", true)

	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(f.Name())

	err = cmd.SendHeartbeats(v, f.Name())
	require.NoError(t, err)

	assert.Equal(t, 0, numCalls)
}

func TestSendHeartbeats_ExtraHeartbeats(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		// check request
		expectedBody, err := ioutil.ReadFile("testdata/api_heartbeats_request_extra_heartbeats_template.json")
		require.NoError(t, err)

		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)

		var entities []struct {
			Entity string `json:"entity"`
		}

		err = json.Unmarshal(body, &entities)
		require.NoError(t, err)

		assert.True(t, strings.HasSuffix(entities[0].Entity, "testdata/main.go"))
		assert.True(t, strings.HasSuffix(entities[1].Entity, "testdata/main.go"))
		assert.True(t, strings.HasSuffix(entities[2].Entity, "testdata/main.py"))

		expectedBodyStr := fmt.Sprintf(
			string(expectedBody),
			entities[0].Entity,
			heartbeat.UserAgent(plugin),
			entities[1].Entity,
			heartbeat.UserAgent(plugin),
			entities[2].Entity,
			heartbeat.UserAgent(plugin),
		)

		assert.JSONEq(t, expectedBodyStr, string(body))

		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)
		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)

		numCalls++
	})

	r, w, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		r.Close()
		w.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = r

	data, err := ioutil.ReadFile("testdata/extra_heartbeats.json")
	require.NoError(t, err)

	go func() {
		_, err := w.Write(data)
		require.NoError(t, err)

		w.Close()
	}()

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", testServerURL)
	v.Set("category", "debugging")
	v.Set("cursorpos", 42)
	v.Set("entity", "testdata/main.go")
	v.Set("entity-type", "file")
	v.Set("extra-heartbeats", true)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("hide-branch-names", "true")
	v.Set("project", "wakatime-cli")
	v.Set("language", "Go")
	v.Set("alternate-language", "Golang")
	v.Set("lineno", 13)
	v.Set("plugin", plugin)
	v.Set("time", 1585598059.1)
	v.Set("timeout", 5)
	v.Set("write", true)

	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(f.Name())

	err = cmd.SendHeartbeats(v, f.Name())
	require.NoError(t, err)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_NonExistingEntity(t *testing.T) {
	logFile, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(logFile.Name())

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", "https://example.org")
	v.Set("entity", "nonexisting")
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("log-file", logFile.Name())

	legacy.SetupLogging(v)

	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(f.Name())

	err = cmd.SendHeartbeats(v, f.Name())
	require.NoError(t, err)

	output, err := ioutil.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "file 'nonexisting' does not exist. ignoring this heartbeat")
}

func TestSendHeartbeats_NonExistingExtraHeartbeatsEntity(t *testing.T) {
	inr, inw, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		inr.Close()
		inw.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = inr

	data, err := ioutil.ReadFile("testdata/extra_heartbeats_nonexisting_entity.json")
	require.NoError(t, err)

	go func() {
		_, err := inw.Write(data)
		require.NoError(t, err)

		inw.Close()
	}()

	logFile, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(logFile.Name())

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", "https://example.org")
	v.Set("entity", "testdata/main.go")
	v.Set("entity-type", "file")
	v.Set("extra-heartbeats", true)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("log-file", logFile.Name())
	v.Set("verbose", true)

	legacy.SetupLogging(v)

	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(f.Name())

	err = cmd.SendHeartbeats(v, f.Name())
	require.NoError(t, err)

	output, err := ioutil.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "file 'nonexisting' does not exist. ignoring this extra heartbeat")
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}
