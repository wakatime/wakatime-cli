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

	rendered, err := summary.RenderToday(s)
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

	rendered, err := summary.RenderToday(s)
	require.NoError(t, err)

	assert.Equal(t, "5hrs 7m", rendered)
}

func TestRenderToday_TotalsByCategory_MultipleCategories(t *testing.T) {
	s := &summary.Summary{
		Total: "6hrs 10m",
		ByCategory: []summary.Category{
			{
				Category: "coding",
				Total:    "5hrs 7m",
			},
			{
				Category: "debugging",
				Total:    "2hrs 3m",
			},
		},
	}

	rendered, err := summary.RenderToday(s)
	require.NoError(t, err)

	assert.Equal(t, "5hrs 7m coding, 2hrs 3m debugging", rendered)
}
