package summary_test

import (
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/summary"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderToday_DayTotal(t *testing.T) {
	var (
		now       = time.Now()
		today     = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		yesterday = today.AddDate(0, 0, -1)
	)

	summaries := []summary.Summary{
		{
			Date:  today,
			Total: "6hrs 10m",
		},
		{
			Date:  yesterday,
			Total: "only today is considered",
		},
	}

	rendered, err := summary.RenderToday(summaries)
	require.NoError(t, err)

	assert.Equal(t, "6hrs 10m", rendered)
}

func TestRenderToday_TotalsByCategory_OneCategory(t *testing.T) {
	var (
		now       = time.Now()
		today     = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		yesterday = today.AddDate(0, 0, -1)
	)

	summaries := []summary.Summary{
		{
			Date:  today,
			Total: "5hrs 7m",
			ByCategory: []summary.Category{
				{
					Category: "coding",
					Total:    "5hrs 7m",
				},
			},
		},
		{
			Date:  yesterday,
			Total: "only today is considered",
		},
	}

	rendered, err := summary.RenderToday(summaries)
	require.NoError(t, err)

	assert.Equal(t, "5hrs 7m", rendered)
}

func TestRenderToday_TotalsByCategory_MultipleCategories(t *testing.T) {
	var (
		now       = time.Now()
		today     = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		yesterday = today.AddDate(0, 0, -1)
	)

	summaries := []summary.Summary{
		{
			Date:  today,
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
		},
		{
			Date:  yesterday,
			Total: "only today is considered",
		},
	}

	rendered, err := summary.RenderToday(summaries)
	require.NoError(t, err)

	assert.Equal(t, "5hrs 7m coding, 2hrs 3m debugging", rendered)
}
