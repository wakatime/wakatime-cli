package heartbeat_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeartbeat_JSON(t *testing.T) {
	h := heartbeat.Heartbeat{
		Branch:         heartbeat.String("heartbeat"),
		Category:       heartbeat.CodingCategory,
		CursorPosition: heartbeat.Int(12),
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "/tmp/main.go",
		EntityType:     heartbeat.FileType,
		IsWrite:        heartbeat.Bool(true),
		Language:       heartbeat.String("golang"),
		LineNumber:     heartbeat.Int(42),
		Lines:          heartbeat.Int(100),
		Project:        heartbeat.String("wakatime"),
		Time:           1585598060.1,
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

func TestHeartbeat_JSON_NilFields(t *testing.T) {
	h := heartbeat.Heartbeat{
		Branch:         nil,
		Category:       heartbeat.CodingCategory,
		CursorPosition: nil,
		Dependencies:   nil,
		Entity:         "/tmp/main.go",
		EntityType:     heartbeat.FileType,
		IsWrite:        nil,
		Language:       nil,
		LineNumber:     nil,
		Lines:          nil,
		Project:        nil,
		Time:           1585598060,
		UserAgent:      "wakatime/13.0.7",
	}

	jsonEncoded, err := json.Marshal(h)
	require.NoError(t, err)

	f, err := os.Open("./testdata/heartbeat_null_fields.json")
	require.NoError(t, err)

	defer f.Close()

	expected, err := ioutil.ReadAll(f)
	require.NoError(t, err)

	assert.JSONEq(t, string(expected), string(jsonEncoded))
}
