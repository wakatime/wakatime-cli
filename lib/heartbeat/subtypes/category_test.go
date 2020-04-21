package subtypes_test

import (
	"encoding/json"
	"testing"

	"github.com/alanhamlett/wakatime-cli/lib/heartbeat/subtypes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var categoryTests = map[string]subtypes.Category{
	"coding":         subtypes.CodingCategory,
	"browsing":       subtypes.BrowsingCategory,
	"building":       subtypes.BuildingCategory,
	"code reviewing": subtypes.CodeReviewingCategory,
	"debugging":      subtypes.DebuggingCategory,
	"designing":      subtypes.DesigningCategory,
	"indexing":       subtypes.IndexingCategory,
	"manual testing": subtypes.ManualTestingCategory,
	"running tests":  subtypes.RunningTestsCategory,
	"writing tests":  subtypes.WritingTestsCategory,
}

func TestCategory_UnmarshalJSON(t *testing.T) {
	for value, category := range categoryTests {
		t.Run(value, func(t *testing.T) {
			var c subtypes.Category
			require.NoError(t, json.Unmarshal([]byte(`"`+value+`"`), &c))

			assert.Equal(t, category, c)
		})
	}
}

func TestCategory_UnmarshalJSON_Invalid(t *testing.T) {
	var c subtypes.Category
	require.Error(t, json.Unmarshal([]byte(`"invalid"`), &c))
}

func TestCategory_MarshalJSON(t *testing.T) {
	for value, category := range categoryTests {
		t.Run(value, func(t *testing.T) {
			data, err := json.Marshal(category)
			require.NoError(t, err)
			assert.JSONEq(t, `"`+value+`"`, string(data))
		})
	}
}

func TestCategory_MarshalJSON_DefaultCategory(t *testing.T) {
	var c subtypes.Category
	data, err := json.Marshal(c)
	require.NoError(t, err)
	assert.JSONEq(t, `"coding"`, string(data))
}

func TestCategory_String(t *testing.T) {
	for value, category := range categoryTests {
		t.Run(value, func(t *testing.T) {
			s := category.String()
			assert.Equal(t, value, s)
		})
	}
}
