package summary_test

import (
	"os"
	"strings"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/output"
	"github.com/wakatime/wakatime-cli/pkg/summary"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderToday(t *testing.T) {
	tests := map[string]struct {
		Output   output.Output
		Expected string
	}{
		"text output": {
			Output:   output.TextOutput,
			Expected: "2 hrs 17 mins Coding, 7 secs Debugging",
		},
		"json output": {
			Output:   output.JSONOutput,
			Expected: readFile(t, "testdata/statusbar_today_simplified.json"),
		},
		"raw json output": {
			Output:   output.RawJSONOutput,
			Expected: readFile(t, "testdata/statusbar_today.json"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			rendered, err := summary.RenderToday(testSummary(), false, test.Output)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, rendered)
		})
	}
}

func TestRenderToday_OneCategory(t *testing.T) {
	s := testSummary()
	s.Data.Categories = s.Data.Categories[:1]

	rendered, err := summary.RenderToday(s, false, output.TextOutput)
	require.NoError(t, err)

	assert.Equal(t, "2 hrs 17 mins", rendered)
}

func TestRenderToday_MultipleCategoriesHidden(t *testing.T) {
	rendered, err := summary.RenderToday(testSummary(), true, output.TextOutput)
	require.NoError(t, err)

	assert.Equal(t, "2 hrs 17 mins", rendered)
}

func readFile(t *testing.T, fp string) string {
	data, err := os.ReadFile(fp)
	require.NoError(t, err)

	return strings.TrimSpace(string(data))
}

func testSummary() *summary.Summary {
	return &summary.Summary{
		CachedAt: "2023-01-29T17:32:05Z",
		Data: summary.Data{
			Categories: []summary.Category{
				{
					Decimal:      "2.28",
					Digital:      "2:17:36",
					Hours:        2,
					Minutes:      17,
					Name:         "Coding",
					Percent:      99.02,
					Seconds:      36,
					Text:         "2 hrs 17 mins",
					TotalSeconds: 8256.598234,
				},
				{
					Decimal:      "0.00",
					Digital:      "0:00:07",
					Hours:        0,
					Minutes:      0,
					Name:         "Debugging",
					Percent:      0.08,
					Seconds:      7,
					Text:         "7 secs",
					TotalSeconds: 7.100772,
				},
			},
			Dependencies: []summary.Dependency{
				{
					Decimal:      "1.25",
					Digital:      "1:15:44",
					Hours:        1,
					Minutes:      15,
					Name:         "strings",
					Percent:      64.82,
					Seconds:      44,
					Text:         "1 hr 15 mins",
					TotalSeconds: 4544.055638,
				},
				{
					Decimal:      "0.82",
					Digital:      "0:49:06",
					Hours:        0,
					Minutes:      49,
					Name:         "io",
					Percent:      35.18,
					Seconds:      6,
					Text:         "49 mins",
					TotalSeconds: 2946.01205,
				},
			},
			Editors: []summary.Editor{
				{
					Decimal:      "2.07",
					Digital:      "2:04:07",
					Hours:        2,
					Minutes:      4,
					Name:         "VS Code",
					Percent:      90.2,
					Seconds:      7,
					Text:         "2 hrs 4 mins",
					TotalSeconds: 7447.112447,
				},
				{
					Decimal:      "0.22",
					Digital:      "0:13:29",
					Hours:        0,
					Minutes:      13,
					Name:         "Zsh-Wakatime-Sobolevn",
					Percent:      9.8,
					Seconds:      29,
					Text:         "13 mins",
					TotalSeconds: 809.485787,
				},
			},
			GrandTotal: summary.GrandTotal{
				Decimal:      "2.28",
				Digital:      "2:17",
				Hours:        2,
				Minutes:      17,
				Text:         "2 hrs 17 mins",
				TotalSeconds: 8256.598234,
			},
			Languages: []summary.Language{
				{
					Decimal:      "1.93",
					Digital:      "1:56:49",
					Hours:        1,
					Minutes:      56,
					Name:         "Go",
					Percent:      86.15,
					Seconds:      49,
					Text:         "1 hr 56 mins",
					TotalSeconds: 7009.317188,
				},
				{
					Decimal:      "0.27",
					Digital:      "0:16:11",
					Hours:        0,
					Minutes:      16,
					Name:         "Other",
					Percent:      13.85,
					Seconds:      11,
					Text:         "16 mins",
					TotalSeconds: 971.489169,
				},
			},
			Machines: []summary.Machine{
				{
					Decimal:       "2.28",
					Digital:       "2:17:36",
					Hours:         2,
					MachineNameID: "370471e8-b6dd-41aa-a94e-d4fb59a7db85",
					Minutes:       17,
					Name:          "WakaMachine",
					Percent:       100.0,
					Seconds:       36,
					Text:          "2 hrs 17 mins",
					TotalSeconds:  8256.598234,
				},
			},
			OperatingSystems: []summary.OperatingSystem{
				{
					Decimal:      "2.28",
					Digital:      "2:17:36",
					Hours:        2,
					Minutes:      17,
					Name:         "Mac",
					Percent:      100.0,
					Seconds:      36,
					Text:         "2 hrs 17 mins",
					TotalSeconds: 8256.598234,
				},
			},
			Projects: []summary.Project{
				{
					Decimal:      "2.05",
					Digital:      "2:03:44",
					Hours:        2,
					Minutes:      3,
					Name:         "wakatime-cli",
					Percent:      97.53,
					Seconds:      44,
					Text:         "2 hrs 3 mins",
					TotalSeconds: 7424.621273,
				},
				{
					Decimal:      "0.05",
					Digital:      "0:03:02",
					Hours:        0,
					Minutes:      3,
					Name:         "Terminal",
					Percent:      2.46,
					Seconds:      2,
					Text:         "3 mins",
					TotalSeconds: 182.934009,
				},
			},
			Range: summary.Range{
				Date:     "2023-01-29",
				End:      "2023-01-30T02:59:59Z",
				Start:    "2023-01-29T03:00:00Z",
				Text:     "Sun Jan 29th 2023",
				Timezone: "America/Sao_Paulo",
			},
		},
		HasTeamFeatures: true,
	}
}
