package goal_test

import (
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/goal"
	"github.com/wakatime/wakatime-cli/pkg/output"

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
			Expected: "2 hrs 1 min",
		},
		"raw json output": {
			Output:   output.RawJSONOutput,
			Expected: readFile(t, "testdata/goal.json"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			rendered, err := goal.RenderToday(testGoal(), test.Output)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, rendered)
		})
	}
}

func readFile(t *testing.T, fp string) string {
	data, err := os.ReadFile(fp)
	require.NoError(t, err)

	return string(data)
}

func testGoal() *goal.Goal {
	return &goal.Goal{
		CachedAt: "2023-01-29T17:28:30Z",
		Data: goal.Data{
			AverageStatus: "success",
			ChartData: []goal.ChartData{
				{
					ActualSeconds:     16708.736808,
					ActualSecondsText: "4 hrs 38 mins",
					GoalSeconds:       3600,
					GoalSecondsText:   "1 hr",
					Range: goal.Range{
						Date:     "2023-01-23",
						End:      "2023-01-24T02:59:59Z",
						Start:    "2023-01-23T03:00:00Z",
						Text:     "Mon Jan 23",
						Timezone: "America/Sao_Paulo",
					},
					RangeStatus:            "success",
					RangeStatusReason:      "coded 4 hrs 38 mins which is 3 hrs 38 mins more than your daily goal",
					RangeStatusReasonShort: "4h 38m (3h 38m more than goal)",
				},
				{
					ActualSeconds:     17518.420996,
					ActualSecondsText: "4 hrs 51 mins",
					GoalSeconds:       3600,
					GoalSecondsText:   "1 hr",
					Range: goal.Range{
						Date:     "2023-01-24",
						End:      "2023-01-25T02:59:59Z",
						Start:    "2023-01-24T03:00:00Z",
						Text:     "Tue Jan 24",
						Timezone: "America/Sao_Paulo",
					},
					RangeStatus:            "success",
					RangeStatusReason:      "coded 4 hrs 51 mins which is 3 hrs 51 mins more than your daily goal",
					RangeStatusReasonShort: "4h 51m (3h 51m more than goal)",
				},
				{
					ActualSeconds:     4530.492416,
					ActualSecondsText: "1 hr 15 mins",
					GoalSeconds:       3600,
					GoalSecondsText:   "1 hr",
					Range: goal.Range{
						Date:     "2023-01-25",
						End:      "2023-01-26T02:59:59Z",
						Start:    "2023-01-25T03:00:00Z",
						Text:     "Wed Jan 25",
						Timezone: "America/Sao_Paulo",
					},
					RangeStatus:            "success",
					RangeStatusReason:      "coded 1 hr 15 mins which is 15 mins more than your daily goal",
					RangeStatusReasonShort: "1h 15m (15m more than goal)",
				},
				{
					ActualSeconds:     14871.537821,
					ActualSecondsText: "4 hrs 7 mins",
					GoalSeconds:       3600,
					GoalSecondsText:   "1 hr",
					Range: goal.Range{
						Date:     "2023-01-26",
						End:      "2023-01-27T02:59:59Z",
						Start:    "2023-01-26T03:00:00Z",
						Text:     "Thu Jan 26",
						Timezone: "America/Sao_Paulo",
					},
					RangeStatus:            "success",
					RangeStatusReason:      "coded 4 hrs 7 mins which is 3 hrs 7 mins more than your daily goal",
					RangeStatusReasonShort: "4h 7m (3h 7m more than goal)",
				},
				{
					ActualSeconds:     17915.08904,
					ActualSecondsText: "4 hrs 58 mins",
					GoalSeconds:       3600,
					GoalSecondsText:   "1 hr",
					Range: goal.Range{
						Date:     "2023-01-27",
						End:      "2023-01-28T02:59:59Z",
						Start:    "2023-01-27T03:00:00Z",
						Text:     "Fri Jan 27",
						Timezone: "America/Sao_Paulo",
					},
					RangeStatus:            "success",
					RangeStatusReason:      "coded 4 hrs 58 mins which is 3 hrs 58 mins more than your daily goal",
					RangeStatusReasonShort: "4h 58m (3h 58m more than goal)",
				},
				{
					ActualSeconds:     10544.828664,
					ActualSecondsText: "2 hrs 55 mins",
					GoalSeconds:       3600,
					GoalSecondsText:   "1 hr",
					Range: goal.Range{
						Date:     "2023-01-28",
						End:      "2023-01-29T02:59:59Z",
						Start:    "2023-01-28T03:00:00Z",
						Text:     "Yesterday",
						Timezone: "America/Sao_Paulo",
					},
					RangeStatus:            "success",
					RangeStatusReason:      "coded 2 hrs 55 mins which is 1 hr 55 mins more than your daily goal",
					RangeStatusReasonShort: "2h 55m (1h 55m more than goal)",
				},
				{
					ActualSeconds:     7288.191208,
					ActualSecondsText: "2 hrs 1 min",
					GoalSeconds:       3600,
					GoalSecondsText:   "1 hr",
					Range: goal.Range{
						Date:     "2023-01-29",
						End:      "2023-01-30T02:59:59Z",
						Start:    "2023-01-29T03:00:00Z",
						Text:     "Today",
						Timezone: "America/Sao_Paulo",
					},
					RangeStatus:            "success",
					RangeStatusReason:      "coded 2 hrs 1 min which is 1 hr 1 min more than your daily goal",
					RangeStatusReasonShort: "2h 1m (1h 1m more than goal)",
				},
			},
			CreatedAt:        "2023-01-29T17:14:49Z",
			CumulativeStatus: "success",
			CustomTitle:      nil,
			Delta:            "day",
			Editors: []string{
				"VS Code",
			},
			ID:                 "0044a592-b3ed-4288-a481-cff56d7a275c",
			IgnoreDays:         []string{},
			IgnoreZeroDays:     true,
			ImproveByPercent:   nil,
			IsCurrentUserOwner: true,
			IsEnabled:          true,
			IsInverse:          false,
			IsSnoozed:          false,
			IsTweeting:         false,
			Languages:          []string{},
			ModifiedAt:         nil,
			Owner: goal.Owner{
				DisplayName: "WakaTime (@wakatime)",
				Email:       nil,
				FullName:    "WakaTime",
				ID:          "fcc2d90e-7665-49f2-a6b1-c49dd0b488cb",
				Photo:       "https://wakatime.com/photo/fcc2d90e-7665-49f2-a6b1-c49dd0b488cb",
				Username:    "wakatime",
			},
			Projects:                []string{},
			RangeText:               "from 2023-01-23 until 2023-01-29",
			Seconds:                 3600,
			SharedWith:              []string{},
			SnoozeUntil:             nil,
			Status:                  "success",
			StatusPercentCalculated: 100,
			Subscribers: []goal.Subscriber{
				{
					DisplayName:    "WakaTime (@wakatime)",
					Email:          nil,
					EmailFrequency: "Daily",
					FullName:       "WakaTime",
					UserID:         "fcc2d90e-7665-49f2-a6b1-c49dd0b488cb",
					Username:       "wakatime",
				},
			},
			Title: "Code 1 hr per day using VS Code",
			Type:  "coding",
		},
	}
}
