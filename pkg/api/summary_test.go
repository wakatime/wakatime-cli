package api_test

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/summary"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSummariesResponse_DayTotal(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/api_summaries_response.json")
	require.NoError(t, err)

	summaries, err := api.ParseSummariesResponse(data)
	require.NoError(t, err)

	assert.Len(t, summaries, 2)
	assert.Contains(t, summaries, summary.Summary{
		Date:  time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		Total: "10 secs",
	})
	assert.Contains(t, summaries, summary.Summary{
		Date:  time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
		Total: "20 secs",
	})
}

func TestParseSummariesResponse_TotalsByCategory(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/api_summaries_by_category_response.json")
	require.NoError(t, err)

	summaries, err := api.ParseSummariesResponse(data)
	require.NoError(t, err)

	assert.Len(t, summaries, 2)
	assert.Contains(t, summaries, summary.Summary{
		Date:  time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		Total: "50 secs",
		ByCategory: []summary.Category{
			{
				Category: "coding",
				Total:    "30 secs",
			},
			{
				Category: "debugging",
				Total:    "20 secs",
			},
		},
	})
	assert.Contains(t, summaries, summary.Summary{
		Date:  time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
		Total: "50 secs",
		ByCategory: []summary.Category{
			{
				Category: "coding",
				Total:    "50 secs",
			},
		},
	})
}
