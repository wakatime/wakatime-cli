package heartbeat_test

import (
	"testing"

	"github.com/alanhamlett/wakatime-cli/lib/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var categoryTests = map[string]heartbeat.Category{
	"unknown":       heartbeat.UnknownCategory,
	"browsing":      heartbeat.BrowsingCategory,
	"building":      heartbeat.BuildingCategory,
	"codeReviewing": heartbeat.CodeReviewingCategory,
	"coding":        heartbeat.CodingCategory,
	"debugging":     heartbeat.DebuggingCategory,
	"designing":     heartbeat.DesigningCategory,
	"indexing":      heartbeat.IndexingCategory,
	"manualTesting": heartbeat.ManualTestingCategory,
	"runningTests":  heartbeat.RunningTestsCategory,
	"writingTests":  heartbeat.WritingTestsCategory,
}

func TestCategory_UnmarshalJSON(t *testing.T) {
	for value, category := range categoryTests {
		t.Run(value, func(t *testing.T) {
			var c heartbeat.Category
			err := c.UnmarshalJSON([]byte(value))
			require.NoError(t, err)

			assert.Equal(t, category, c)
		})
	}
}

func TestCategory_UnmarshalJSON_Invalid(t *testing.T) {
	var c heartbeat.Category
	err := c.UnmarshalJSON([]byte("invalid"))
	require.Error(t, err)
}

func TestCategory_MarshalJSON(t *testing.T) {
	for value, category := range categoryTests {
		t.Run(value, func(t *testing.T) {
			val, err := category.MarshalJSON()
			require.NoError(t, err)

			assert.Equal(t, value, string(val))
		})
	}
}

func TestCategory_String(t *testing.T) {
	for value, category := range categoryTests {
		t.Run(value, func(t *testing.T) {
			s := category.String()
			assert.Equal(t, value, s)
		})
	}
}
