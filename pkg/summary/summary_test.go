package summary_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/summary"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderToday_DayTotal(t *testing.T) {
	s := &summary.Summary{
		Total: "6hrs 10m",
	}

	rendered, err := summary.RenderToday(s, false)
	require.NoError(t, err)

	assert.Equal(t, "6hrs 10m", rendered)
}

func TestRenderToday_TotalsByCategory_OneCategory(t *testing.T) {
	s := &summary.Summary{
		Total: "5hrs 7m",
		ByCategory: []summary.Category{
			{
				Category: "coding",
				Total:    "5hrs 7m",
			},
		},
	}

	rendered, err := summary.RenderToday(s, false)
	require.NoError(t, err)

	assert.Equal(t, "5hrs 7m", rendered)
}

func TestRenderToday_TotalsByCategory_MultipleCategories(t *testing.T) {
	s := &summary.Summary{
		Total: "6 hrs 10 mins",
		ByCategory: []summary.Category{
			{
				Category: "coding",
				Total:    "5 hrs 7 mins",
			},
			{
				Category: "debugging",
				Total:    "2 hrs 3 mins",
			},
		},
	}

	rendered, err := summary.RenderToday(s, false)
	require.NoError(t, err)

	assert.Equal(t, "5 hrs 7 mins coding, 2 hrs 3 mins debugging", rendered)
}

func TestRenderToday_TotalsByCategory_MultipleCategoriesHidden(t *testing.T) {
	s := &summary.Summary{
		Total: "6 hrs 10 mins",
		ByCategory: []summary.Category{
			{
				Category: "coding",
				Total:    "5 hrs 7 mins",
			},
			{
				Category: "debugging",
				Total:    "2 hrs 3 mins",
			},
		},
	}

	rendered, err := summary.RenderToday(s, true)
	require.NoError(t, err)

	assert.Equal(t, "6 hrs 10 mins", rendered)
}
