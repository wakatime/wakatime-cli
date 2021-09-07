//go:build integration

package main_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/exitcode"
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
	testVersion       = "<local-build>"
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
	runCmd(exec.Command("cp", "./testdata/main.go", tmpDir))

	testSendHeartbeats(t, filepath.Join(tmpDir, "main.go"), "")
}

func testSendHeartbeats(t *testing.T, entity, project string) {
	apiURL, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
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

	offlineQueueFile, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(offlineQueueFile.Name())

	tmpFile, err := ioutil.TempFile(os.TempDir(), "wakatime.cfg")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	runWakatimeCli(
		t,
		"--api-url", apiURL,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", tmpFile.Name(),
		"--entity", entity,
		"--cursorpos", "12",
		"--offline-queue-file", offlineQueueFile.Name(),
		"--lineno", "42",
		"--lines-in-file", "100",
		"--time", "1585598059",
		"--hide-branch-names", ".*",
		"--write",
		"--verbose",
	)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_Err(t *testing.T) {
	apiURL, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
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

		entityPath, err := realpath.Realpath("testdata/main.go")
		require.NoError(t, err)

		entityPath = strings.ReplaceAll(entityPath, `\`, `/`)
		expectedBody := fmt.Sprintf(
			string(expectedBodyTpl),
			entityPath,
			"wakatime-cli",
			heartbeat.UserAgentUnknownPlugin(),
		)

		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)

		assert.JSONEq(t, string(expectedBody), string(body))

		// write response
		w.WriteHeader(http.StatusBadRequest)
	})

	offlineQueueFile, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(offlineQueueFile.Name())

	tmpFile, err := ioutil.TempFile(os.TempDir(), "wakatime.cfg")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	out := runWakatimeCliExpectErr(
		t,
		exitcode.ErrDefault,
		"--api-url", apiURL,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", tmpFile.Name(),
		"--entity", "testdata/main.go",
		"--cursorpos", "12",
		"--offline-queue-file", offlineQueueFile.Name(),
		"--lineno", "42",
		"--lines-in-file", "100",
		"--time", "1585598059",
		"--hide-branch-names", ".*",
		"--write",
		"--verbose",
	)

	assert.Empty(t, out)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestTodayGoal(t *testing.T) {
	apiURL, router, close := setupTestServer()
	defer close()

	var numCalls int

	tmpFile, err := ioutil.TempFile(os.TempDir(), "wakatime.cfg")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	router.HandleFunc("/users/current/goals/11111111-1111-4111-8111-111111111111",
		func(w http.ResponseWriter, req *http.Request) {
			numCalls++

			// check request
			assert.Equal(t, http.MethodGet, req.Method)
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

	out := runWakatimeCli(
		t,
		"--api-url", apiURL,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", tmpFile.Name(),
		"--today-goal", "11111111-1111-4111-8111-111111111111",
		"--verbose",
	)

	assert.Equal(t, "3 hrs 23 mins\n", out)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestTodaySummary(t *testing.T) {
	apiURL, router, close := setupTestServer()
	defer close()

	var numCalls int

	tmpFile, err := ioutil.TempFile(os.TempDir(), "wakatime.cfg")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	router.HandleFunc("/users/current/statusbar/today", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check request
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.Equal(t, []string{heartbeat.UserAgentUnknownPlugin()}, req.Header["User-Agent"])

		// write response
		f, err := os.Open("testdata/api_statusbar_today_response.json")
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	out := runWakatimeCli(
		t,
		"--api-url", apiURL,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", tmpFile.Name(),
		"--today",
		"--verbose",
	)

	assert.Equal(t, "20 secs\n", out)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestOfflineCountEmpty(t *testing.T) {
	offlineQueueFile, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(offlineQueueFile.Name())

	out := runWakatimeCli(
		t,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--offline-queue-file", offlineQueueFile.Name(),
		"--offline-count",
		"--verbose",
	)

	assert.Equal(t, "0\n", out)
}

func TestOfflineCountWithOneHeartbeat(t *testing.T) {
	apiURL, router, close := setupTestServer()
	defer close()

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := io.Copy(w, strings.NewReader("500 error test"))
		require.NoError(t, err)
	})

	offlineQueueFile, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(offlineQueueFile.Name())

	tmpFile, err := ioutil.TempFile(os.TempDir(), "wakatime.cfg")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	out := runWakatimeCliExpectErr(
		t,
		exitcode.ErrAPI,
		"--api-url", apiURL,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", tmpFile.Name(),
		"--entity", "testdata/main.go",
		"--cursorpos", "12",
		"--offline-queue-file", offlineQueueFile.Name(),
		"--lineno", "42",
		"--lines-in-file", "100",
		"--time", "1585598059",
		"--hide-branch-names", ".*",
		"--log-to-stdout",
		"--write",
		"--verbose",
	)

	assert.Empty(t, out)

	out = runWakatimeCli(
		t,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--offline-queue-file", offlineQueueFile.Name(),
		"--offline-count",
		"--verbose",
	)

	assert.Equal(t, "1\n", out)
}

func TestUseragent(t *testing.T) {
	out := runWakatimeCli(t, "--useragent")
	assert.Equal(t, fmt.Sprintf("%s\n", heartbeat.UserAgentUnknownPlugin()), out)
}

func TestUseragentWithPlugin(t *testing.T) {
	out := runWakatimeCli(t, "--useragent", "--plugin", "Wakatime/1.0.4")

	assert.Equal(t, fmt.Sprintf("%s\n", heartbeat.UserAgent("Wakatime/1.0.4")), out)
}

func TestVersion(t *testing.T) {
	out := runWakatimeCli(t, "--version")

	assert.Equal(t, "<local-build>\n", out)
}

func TestVersionVerbose(t *testing.T) {
	out := runWakatimeCli(t, "--version", "--verbose")

	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(
		"wakatime-cli\n  Version: <local-build>\n  Commit: [0-9a-f]{7}\n  Built: [0-9-:T]{19} UTC\n  OS/Arch: %s/%s\n",
		runtime.GOOS,
		runtime.GOARCH,
	)), out)
}

func runWakatimeCli(t *testing.T, args ...string) string {
	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer func() {
		f.Close()
		data, err := ioutil.ReadFile(f.Name())
		require.NoError(t, err)

		fmt.Printf("logs: %s\n", string(data))

		os.Remove(f.Name())
	}()

	args = append([]string{"--log-file", f.Name()}, args...)

	return runCmd(exec.Command(binaryPath(t), args...)) // #nosec G204
}

func runWakatimeCliExpectErr(t *testing.T, exitcode int, args ...string) string {
	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer func() {
		f.Close()
		data, err := ioutil.ReadFile(f.Name())
		require.NoError(t, err)

		fmt.Printf("logs: %s\n", string(data))

		os.Remove(f.Name())
	}()

	args = append([]string{"--log-file", f.Name()}, args...)

	stdout, code := runCmdExpectErr(exec.Command(binaryPath(t), args...)) // #nosec G204

	assert.Equal(t, exitcode, code)

	return stdout
}

func runCmd(cmd *exec.Cmd) string {
	fmt.Println(cmd.String())

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	version.Version = testVersion

	err := cmd.Run()
	if err != nil {
		fmt.Println(stdout.String())
		fmt.Println(stderr.String())
		fmt.Printf("failed to run command %s\n", cmd)
		os.Exit(1)
	}

	return stdout.String()
}

func runCmdExpectErr(cmd *exec.Cmd) (string, int) {
	fmt.Println(cmd.String())

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	version.Version = testVersion

	err := cmd.Run()
	if err == nil {
		fmt.Println(stdout.String())
		fmt.Println(stderr.String())
		fmt.Printf("ran command successfully, but was expecting error: %s\n", cmd)
		os.Exit(1)
	}

	if exitcode, ok := err.(*exec.ExitError); ok {
		return stdout.String(), exitcode.ExitCode()
	}

	return stdout.String(), -1
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}
