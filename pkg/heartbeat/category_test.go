package heartbeat_test

import (
	"encoding/json"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func categoryTests() map[string]heartbeat.Category {
	return map[string]heartbeat.Category{
		"browsing":       heartbeat.BrowsingCategory,
		"building":       heartbeat.BuildingCategory,
		"code reviewing": heartbeat.CodeReviewingCategory,
		"coding":         heartbeat.CodingCategory,
		"communicating":  heartbeat.CommunicatingCategory,
		"debugging":      heartbeat.DebuggingCategory,
		"designing":      heartbeat.DesigningCategory,
		"indexing":       heartbeat.IndexingCategory,
		"learning":       heartbeat.LearningCategory,
		"manual testing": heartbeat.ManualTestingCategory,
		"meeting":        heartbeat.MeetingCategory,
		"planning":       heartbeat.PlanningCategory,
		"researching":    heartbeat.ResearchingCategory,
		"running tests":  heartbeat.RunningTestsCategory,
		"translating":    heartbeat.TranslatingCategory,
		"writing docs":   heartbeat.WritingDocsCategory,
		"writing tests":  heartbeat.WritingTestsCategory,
	}
}

func TestParseCategory(t *testing.T) {
	for value, category := range categoryTests() {
		t.Run(value, func(t *testing.T) {
			parsed, err := heartbeat.ParseCategory(value)
			require.NoError(t, err)

			assert.Equal(t, category, parsed)
		})
	}
}

func TestParseCategory_Invalid(t *testing.T) {
	_, err := heartbeat.ParseCategory("invalid")
	require.Error(t, err)
}

func TestCategory_UnmarshalJSON(t *testing.T) {
	for value, category := range categoryTests() {
		t.Run(value, func(t *testing.T) {
			var c heartbeat.Category
			require.NoError(t, json.Unmarshal([]byte(`"`+value+`"`), &c))

			assert.Equal(t, category, c)
		})
	}
}

func TestCategory_UnmarshalJSON_Invalid(t *testing.T) {
	var c heartbeat.Category

	require.Error(t, json.Unmarshal([]byte(`"invalid"`), &c))
}

func TestCategory_MarshalJSON(t *testing.T) {
	for value, category := range categoryTests() {
		t.Run(value, func(t *testing.T) {
			data, err := json.Marshal(category)
			require.NoError(t, err)
			assert.JSONEq(t, `"`+value+`"`, string(data))
		})
	}
}

func TestCategory_MarshalJSON_DefaultCategory(t *testing.T) {
	var c heartbeat.Category
	data, err := json.Marshal(c)
	require.NoError(t, err)
	assert.JSONEq(t, `"coding"`, string(data))
}

func TestCategory_String(t *testing.T) {
	for value, category := range categoryTests() {
		t.Run(value, func(t *testing.T) {
			s := category.String()
			assert.Equal(t, value, s)
		})
	}
}
