package heartbeat_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/alanhamlett/wakatime-cli/lib/heartbeat"
	"github.com/alanhamlett/wakatime-cli/lib/heartbeat/subtypes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeartbeat_ID(t *testing.T) {
	h := heartbeat.Heartbeat{
		Branch:     "heartbeat",
		Category:   subtypes.CodingCategory,
		Entity:     "/tmp/main.go",
		EntityType: subtypes.FileType,
		IsWrite:    true,
		Project:    "wakatime",
		Time:       1585598060,
	}
	assert.Equal(t, "1585598060-file-coding-wakatime-heartbeat-/tmp/main.go-true", h.ID())
}

func TestHeartbeat_JSON(t *testing.T) {
	h := heartbeat.Heartbeat{
		Branch:         "heartbeat",
		Category:       subtypes.CodingCategory,
		CursorPosition: 12,
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "/tmp/main.go",
		EntityType:     subtypes.FileType,
		IsWrite:        true,
		Language:       "golang",
		LineNumber:     42,
		Lines:          100,
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
