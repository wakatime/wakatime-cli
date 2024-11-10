package heartbeat_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/version"

	"github.com/matishsiao/goInfo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	h := heartbeat.New(
		"feature/branch",
		heartbeat.CodingCategory,
		heartbeat.PointerTo(12),
		"testdata/main.go",
		heartbeat.FileType,
		true,
		heartbeat.PointerTo(true),
		heartbeat.PointerTo("Go"),
		"Golang",
		heartbeat.PointerTo(2),
		heartbeat.PointerTo(3),
		heartbeat.PointerTo(42),
		nil,
		"/path/to/file",
		"billing",
		false,
		"pci",
		"/custom-path",
		1592868313.541149,
		"wakatime/13.0.7",
	)

	assert.True(t, strings.HasSuffix(h.Entity, "testdata/main.go"))

	assert.Equal(t, heartbeat.Heartbeat{
		BranchAlternate:     "feature/branch",
		Category:            heartbeat.CodingCategory,
		CursorPosition:      heartbeat.PointerTo(12),
		EntityType:          heartbeat.FileType,
		IsUnsavedEntity:     true,
		IsWrite:             heartbeat.PointerTo(true),
		Language:            heartbeat.PointerTo("Go"),
		LanguageAlternate:   "Golang",
		LineAdditions:       heartbeat.PointerTo(2),
		LineDeletions:       heartbeat.PointerTo(3),
		LineNumber:          heartbeat.PointerTo(42),
		LocalFile:           "/path/to/file",
		ProjectAlternate:    "billing",
		ProjectOverride:     "pci",
		ProjectPathOverride: "/custom-path",
		Time:                1592868313.541149,
		UserAgent:           "wakatime/13.0.7",
		Entity:              h.Entity,
	}, h)
}

func TestHeartbeat_ID(t *testing.T) {
	h := heartbeat.Heartbeat{
		Branch:     heartbeat.PointerTo("heartbeat"),
		Category:   heartbeat.CodingCategory,
		Entity:     "/tmp/main.go",
		EntityType: heartbeat.FileType,
		IsWrite:    heartbeat.PointerTo(true),
		Project:    heartbeat.PointerTo("wakatime"),
		Time:       1592868313.541149,
	}
	assert.Equal(t, "1592868313.541149-nil-file-coding-wakatime-heartbeat-/tmp/main.go-true", h.ID())
}

func TestHeartbeat_ID_NilFields(t *testing.T) {
	h := heartbeat.Heartbeat{
		Category:   heartbeat.CodingCategory,
		Entity:     "/tmp/main.go",
		EntityType: heartbeat.FileType,
		Time:       1592868313.541149,
	}
	assert.Equal(t, "1592868313.541149-nil-file-coding-unset-unset-/tmp/main.go-false", h.ID())
}

func TestHeartbeat_JSON(t *testing.T) {
	h := heartbeat.Heartbeat{
		Branch:         heartbeat.PointerTo("heartbeat"),
		Category:       heartbeat.CodingCategory,
		CursorPosition: heartbeat.PointerTo(12),
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "/tmp/main.go",
		EntityType:     heartbeat.FileType,
		IsWrite:        heartbeat.PointerTo(true),
		Language:       heartbeat.PointerTo("Go"),
		LineAdditions:  heartbeat.PointerTo(123),
		LineDeletions:  heartbeat.PointerTo(456),
		LineNumber:     heartbeat.PointerTo(42),
		Lines:          heartbeat.PointerTo(100),
		Project:        heartbeat.PointerTo("wakatime"),
		Time:           1585598060.1,
		UserAgent:      "wakatime/13.0.7",
	}

	jsonEncoded, err := json.Marshal(h)
	require.NoError(t, err)

	f, err := os.Open("./testdata/heartbeat.json")
	require.NoError(t, err)

	defer f.Close()

	expected, err := io.ReadAll(f)
	require.NoError(t, err)

	assert.JSONEq(t, string(expected), string(jsonEncoded))
}

func TestHeartbeat_JSON_NilFields(t *testing.T) {
	h := heartbeat.Heartbeat{
		Category:   heartbeat.CodingCategory,
		Entity:     "/tmp/main.go",
		EntityType: heartbeat.FileType,
		Time:       1585598060,
		UserAgent:  "wakatime/13.0.7",
	}

	jsonEncoded, err := json.Marshal(h)
	require.NoError(t, err)

	f, err := os.Open("./testdata/heartbeat_null_fields.json")
	require.NoError(t, err)

	defer f.Close()

	expected, err := io.ReadAll(f)
	require.NoError(t, err)

	assert.JSONEq(t, string(expected), string(jsonEncoded))
}

func TestNewHandle(t *testing.T) {
	sender := mockSender{
		SendHeartbeatsFn: func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			assert.Equal(t, []heartbeat.Heartbeat{
				{
					Branch:     heartbeat.PointerTo("test"),
					Category:   heartbeat.CodingCategory,
					Entity:     "/tmp/main.go",
					EntityType: heartbeat.FileType,
					Time:       1585598060,
					UserAgent:  "wakatime/13.0.7",
				},
			}, hh)
			return []heartbeat.Result{
				{
					Status:    201,
					Heartbeat: heartbeat.Heartbeat{},
				},
			}, nil
		},
	}

	opts := []heartbeat.HandleOption{
		func(next heartbeat.Handle) heartbeat.Handle {
			return func(ctx context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
				for i := range hh {
					hh[i].Branch = heartbeat.PointerTo("test")
				}

				return next(ctx, hh)
			}
		},
	}

	handle := heartbeat.NewHandle(&sender, opts...)
	_, err := handle(context.Background(), []heartbeat.Heartbeat{
		{
			Category:   heartbeat.CodingCategory,
			Entity:     "/tmp/main.go",
			EntityType: heartbeat.FileType,
			Time:       1585598060,
			UserAgent:  "wakatime/13.0.7",
		},
	})
	require.NoError(t, err)
}

func TestUserAgentUnknownPlugin(t *testing.T) {
	info, err := goInfo.GetInfo()
	require.NoError(t, err)

	expected := fmt.Sprintf(
		"wakatime/%s (%s-%s-%s) %s Unknown/0",
		version.Version,
		runtime.GOOS,
		info.Core,
		info.Platform,
		runtime.Version(),
	)

	assert.Equal(t, expected, heartbeat.UserAgent(context.Background(), ""))
}

func TestUserAgent(t *testing.T) {
	info, err := goInfo.GetInfo()
	require.NoError(t, err)

	expected := fmt.Sprintf(
		"wakatime/%s (%s-%s-%s) %s testplugin",
		version.Version,
		runtime.GOOS,
		info.Core,
		info.Platform,
		runtime.Version(),
	)

	assert.Equal(t, expected, heartbeat.UserAgent(context.Background(), "testplugin"))
}

func TestRemoteAddressRegex(t *testing.T) {
	tests := map[string]struct {
		Heartbeat heartbeat.Heartbeat
		Expected  bool
	}{
		"ssh full path": {
			Heartbeat: heartbeat.Heartbeat{Entity: "ssh://user:1234@192.168.1.2/home/pi/unicorn-hat/examples/ascii_pic.py"},
			Expected:  true,
		},
		"sftp full path": {
			Heartbeat: heartbeat.Heartbeat{Entity: "sftp://user:1234@192.168.1.2/home/pi/unicorn-hat/examples/ascii_pic.py"},
			Expected:  true,
		},
		"without path": {
			Heartbeat: heartbeat.Heartbeat{Entity: "ssh://user:1234@192.168.1.2"},
			Expected:  true,
		},
		"invalid ftp": {
			Heartbeat: heartbeat.Heartbeat{Entity: "ftp://user:1234@192.168.1.2"},
			Expected:  false,
		},
		"invalid": {
			Heartbeat: heartbeat.Heartbeat{Entity: "http://192.168.1.2"},
			Expected:  false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := test.Heartbeat.IsRemote()

			assert.Equal(t, test.Expected, result)
		})
	}
}

type mockSender struct {
	SendHeartbeatsFn        func(context.Context, []heartbeat.Heartbeat) ([]heartbeat.Result, error)
	SendHeartbeatsFnInvoked bool
}

func (m *mockSender) SendHeartbeats(ctx context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
	m.SendHeartbeatsFnInvoked = true
	return m.SendHeartbeatsFn(ctx, hh)
}
