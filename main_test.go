// +build integration

package main_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/version"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yookoala/realpath"
)

const (
	binaryPathDarwin  = "./build/wakatime-cli-darwin-amd64"
	binaryPathLinux   = "./build/wakatime-cli-linux-amd64"
	binaryPathWindows = "./build/wakatime-cli-windows-amd64.exe"
	testVersion       = "v0.0.1-test"
)

func binaryPath(t *testing.T) string {
	switch runtime.GOOS {
	case "darwin":
		return binaryPathDarwin
	case "linux":
		return binaryPathLinux
	case "windows":
		return binaryPathWindows
	default:
		t.Fatalf("OS %q not supported", runtime.GOOS)
		return ""
	}
}

func TestSendHeartbeats(t *testing.T) {
	testSendHeartbeats(t, "testdata/main.go", "wakatime-cli")
}

func TestSendHeartbeats_EntityFileInTempDir(t *testing.T) {
	tmpDir := t.TempDir()
	run(exec.Command("cp", "./testdata/main.go", tmpDir))

	testSendHeartbeats(t, filepath.Join(tmpDir, "main.go"), "")
}

func testSendHeartbeats(t *testing.T, entity, project string) {
	apiUrl, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/v1/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check headers
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.Equal(t, []string{heartbeat.UserAgentUnknownPlugin()}, req.Header["User-Agent"])

		// check body
		expectedBodyTpl, err := ioutil.ReadFile("testdata/api_heartbeats_request.json.tpl")
		require.NoError(t, err)

		entityPath, err := realpath.Realpath(entity)
		require.NoError(t, err)

		entityPath = strings.ReplaceAll(entityPath, `\`, `/`)
		expectedBody := fmt.Sprintf(
			string(expectedBodyTpl),
			entityPath,
			project,
			heartbeat.UserAgentUnknownPlugin(),
		)

		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)

		assert.JSONEq(t, string(expectedBody), string(body))

		// write response
		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)

		w.WriteHeader(http.StatusCreated)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	version.Version = testVersion

	cmd := exec.Command(
		binaryPath(t),
		"--api-url", apiUrl,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", "testdata/wakatime.cfg",
		"--entity", entity,
		"--cursorpos", "12",
		"--lineno", "42",
		"--lines-in-file", "100",
		"--time", "1585598059",
		"--hide-branch-names", ".*",
		"--log-to-stdout",
		"--write",
		"--verbose",
	)

	run(cmd)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestTodayGoal(t *testing.T) {
	apiUrl, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/v1/users/current/goals/11111111-1111-4111-8111-111111111111", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check request
		assert.Equal(t, http.MethodGet, req.Method)
		t.Logf("%#v\n", req.Header)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.Equal(t, []string{heartbeat.UserAgentUnknownPlugin()}, req.Header["User-Agent"])

		// write response
		f, err := os.Open("testdata/api_goals_id_response.json")
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	version.Version = testVersion

	out := run(exec.Command(
		binaryPath(t),
		"--api-url", apiUrl,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", "testdata/wakatime.cfg",
		"--today-goal", "11111111-1111-4111-8111-111111111111",
		"--verbose",
	))

	assert.Equal(t, "3 hrs 23 mins\n", out)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestTodaySummary(t *testing.T) {
	apiUrl, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/v1/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check request
		assert.Equal(t, http.MethodGet, req.Method)
		t.Logf("%#v\n", req.Header)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.Equal(t, []string{heartbeat.UserAgentUnknownPlugin()}, req.Header["User-Agent"])

		values, err := url.ParseQuery(req.URL.RawQuery)
		require.NoError(t, err)

		today := time.Now().Format("2006-01-02")

		assert.Equal(t, url.Values(map[string][]string{
			"start": {today},
			"end":   {today},
		}), values)

		// write response
		responseBodyTpl, err := ioutil.ReadFile("testdata/api_summaries_response.json.tpl")
		require.NoError(t, err)

		responseBody := fmt.Sprintf(string(responseBodyTpl), today)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(responseBody))
		require.NoError(t, err)
	})

	version.Version = testVersion

	out := run(exec.Command(
		binaryPath(t),
		"--api-url", apiUrl,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", "testdata/wakatime.cfg",
		"--today",
		"--verbose",
	))

	assert.Equal(t, "10 secs\n", out)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestUseragent(t *testing.T) {
	out := run(exec.Command(binaryPath(t), "--useragent"))

	assert.Equal(t, fmt.Sprintf("%s\n", heartbeat.UserAgentUnknownPlugin()), out)
}

func TestUseragentWithPlugin(t *testing.T) {
	out := run(exec.Command(binaryPath(t), "--useragent", "--plugin", "Wakatime/1.0.4"))

	assert.Equal(t, fmt.Sprintf("%s\n", heartbeat.UserAgent("Wakatime/1.0.4")), out)
}

func TestVersion(t *testing.T) {
	out := run(exec.Command(binaryPath(t), "--version"))

	assert.Equal(t, "v0.0.1-test\n", out)
}

func TestVersionVerbose(t *testing.T) {
	out := run(exec.Command(binaryPath(t), "--version", "--verbose"))

	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(
		"wakatime-cli\n  Version: v0.0.1-test\n  Commit: [0-9a-f]{7}\n  Built: [0-9-:TZ]{20}\n  OS/Arch: %s/%s\n",
		runtime.GOOS,
		runtime.GOARCH,
	)), out)
}

func run(cmd *exec.Cmd) string {
	fmt.Println(cmd.String())
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(stdout.String())
		fmt.Println(stderr.String())
		fmt.Printf("failed to run command %s\n", cmd)
		os.Exit(1)
	}

	return stdout.String()
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}
