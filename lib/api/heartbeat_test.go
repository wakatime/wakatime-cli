package api_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/alanhamlett/wakatime-cli/lib/api"
	"github.com/alanhamlett/wakatime-cli/lib/heartbeat/subtypes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeartbeat_JSON(t *testing.T) {
	h := api.Heartbeat{
		Branch:         String("heartbeat"),
		Category:       subtypes.CodingCategory,
		CursorPosition: Int(12),
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "/tmp/main.go",
		EntityType:     subtypes.FileType,
		IsWrite:        true,
		Language:       "golang",
		LineNumber:     Int(42),
		Lines:          Int(100),
		Project:        "wakatime",
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}

	jsonEncoded, err := json.Marshal(h)
	require.NoError(t, err)

	f, err := os.Open("./testdata/heartbeat.json")
	require.NoError(t, err)
	defer f.Close()

	expected, err := ioutil.ReadAll(f)
	require.NoError(t, err)

	assert.JSONEq(t, string(expected), string(jsonEncoded))
}

func TestHeartbeat_JSON_Sanitized(t *testing.T) {
	h := api.Heartbeat{
		Branch:         nil,
		Category:       subtypes.CodingCategory,
		CursorPosition: nil,
		Dependencies:   nil,
		Entity:         "HIDDEN.go",
		EntityType:     subtypes.FileType,
		IsWrite:        true,
		Language:       "golang",
		LineNumber:     nil,
		Lines:          nil,
		Project:        "wakatime",
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}

	jsonEncoded, err := json.Marshal(h)
	require.NoError(t, err)

	f, err := os.Open("./testdata/heartbeat_sanitized.json")
	require.NoError(t, err)
	defer f.Close()

	expected, err := ioutil.ReadAll(f)
	require.NoError(t, err)

	assert.JSONEq(t, string(expected), string(jsonEncoded))
}

// Int returns a pointer to the int value passed in.
func Int(v int) *int {
	return &v
}

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}
