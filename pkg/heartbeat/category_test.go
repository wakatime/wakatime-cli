package heartbeat_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func categoryTests() map[string]heartbeat.Category {
	return map[string]heartbeat.Category{
		"advising":       heartbeat.AdvisingCategory,
		"ai coding":      heartbeat.AICodingCategory,
		"browsing":       heartbeat.BrowsingCategory,
		"building":       heartbeat.BuildingCategory,
		"code reviewing": heartbeat.CodeReviewingCategory,
		"communicating":  heartbeat.CommunicatingCategory,
		"debugging":      heartbeat.DebuggingCategory,
		"designing":      heartbeat.DesigningCategory,
		"indexing":       heartbeat.IndexingCategory,
		"learning":       heartbeat.LearningCategory,
		"manual testing": heartbeat.ManualTestingCategory,
		"meeting":        heartbeat.MeetingCategory,
		"notes":          heartbeat.NotesCategory,
		"planning":       heartbeat.PlanningCategory,
		"researching":    heartbeat.ResearchingCategory,
		"running tests":  heartbeat.RunningTestsCategory,
		"supporting":     heartbeat.SupportingCategory,
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

	assert.Error(t, json.Unmarshal([]byte(`"invalid"`), &c))
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

func TestCategory_MarshalJSON_UndefinedCategory(t *testing.T) {
	data, err := json.Marshal(heartbeat.UndefinedCategory)
	require.NoError(t, err)

	assert.JSONEq(t, `null`, string(data))
}

func TestCategory_String(t *testing.T) {
	for value, category := range categoryTests() {
		t.Run(value, func(t *testing.T) {
			s := category.String()
			assert.Equal(t, value, s)
		})
	}
}

func TestWithCategory(t *testing.T) {
	opt := heartbeat.WithCategory()

	handle := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, heartbeat.UndefinedCategory.String(), hh[0].Category)
		assert.Equal(t, heartbeat.UndefinedCategory.String(), hh[1].Category)
		assert.Equal(t, heartbeat.WritingTestsCategory.String(), hh[2].Category)
		assert.Equal(t, heartbeat.WritingTestsCategory.String(), hh[3].Category)
		assert.Equal(t, heartbeat.WritingTestsCategory.String(), hh[4].Category)
		assert.Equal(t, heartbeat.WritingTestsCategory.String(), hh[5].Category)
		assert.Equal(t, heartbeat.WritingTestsCategory.String(), hh[6].Category)
		assert.Equal(t, heartbeat.WritingTestsCategory.String(), hh[7].Category)
		assert.Equal(t, heartbeat.WritingTestsCategory.String(), hh[8].Category)
		assert.Equal(t, heartbeat.WritingDocsCategory.String(), hh[9].Category)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := handle(t.Context(), []heartbeat.Heartbeat{
		{
			Entity: "/foo/file.go",
		},
		{
			Entity:   "/foo/file.go",
			Category: heartbeat.CodingCategory.String(), // coding category changes to empty string (undefined category)
		},
		{
			Entity: "/foo/foo_test.go",
		},
		{
			Entity: "/foo/spec/file.rb",
		},
		{
			Entity: "/foo/specs/file.rb",
		},
		{
			Entity: "/foo/test/file.py",
		},
		{
			Entity: "/foo/tests/file.py",
		},
		{
			Entity: "/foo/testdata/file.py",
		},
		{
			Entity: "/foo/testdata/file.md",
		},
		{
			Entity: "/foo/file.md",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestWithCategory_NotFileType(t *testing.T) {
	opt := heartbeat.WithCategory()

	handle := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, heartbeat.DebuggingCategory.String(), hh[0].Category)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := handle(t.Context(), []heartbeat.Heartbeat{
		{
			Entity:     "/foo/file.go",
			EntityType: heartbeat.AppType,
			Category:   heartbeat.DebuggingCategory.String(),
		},
	})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestWithCategory_CodingCategory(t *testing.T) {
	opt := heartbeat.WithCategory()

	handle := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Empty(t, hh[0].Category)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := handle(t.Context(), []heartbeat.Heartbeat{
		{
			Entity:   "/foo/file.go",
			Category: heartbeat.CodingCategory.String(),
		},
	})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}
