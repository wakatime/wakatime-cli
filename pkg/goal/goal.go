package goal

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/wakatime/wakatime-cli/pkg/output"
)

type (
	// Goal represents a goal.
	Goal struct {
		CachedAt string `json:"cached_at"`
		Data     Data   `json:"data"`
	}

	// Range represents the time range of a goal.
	Range struct {
		Date     string `json:"date"`
		End      string `json:"end"`
		Start    string `json:"start"`
		Text     string `json:"text"`
		Timezone string `json:"timezone"`
	}

	// ChartData represents the chart data of a goal.
	ChartData struct {
		ActualSeconds          float64 `json:"actual_seconds"`
		ActualSecondsText      string  `json:"actual_seconds_text"`
		GoalSeconds            int     `json:"goal_seconds"`
		GoalSecondsText        string  `json:"goal_seconds_text"`
		Range                  Range   `json:"range"`
		RangeStatus            string  `json:"range_status"`
		RangeStatusReason      string  `json:"range_status_reason"`
		RangeStatusReasonShort string  `json:"range_status_reason_short"`
	}

	// Owner represents the owner of a goal.
	Owner struct {
		DisplayName string  `json:"display_name"`
		Email       *string `json:"email"`
		FullName    string  `json:"full_name"`
		ID          string  `json:"id"`
		Photo       string  `json:"photo"`
		Username    string  `json:"username"`
	}

	// Subscriber represents a subscriber of a goal.
	Subscriber struct {
		DisplayName    string  `json:"display_name"`
		Email          *string `json:"email"`
		EmailFrequency string  `json:"email_frequency"`
		FullName       string  `json:"full_name"`
		UserID         string  `json:"user_id"`
		Username       string  `json:"username"`
	}

	// Data represents the data of a goal.
	Data struct {
		AverageStatus           string       `json:"average_status"`
		ChartData               []ChartData  `json:"chart_data"`
		CreatedAt               string       `json:"created_at"`
		CumulativeStatus        string       `json:"cumulative_status"`
		CustomTitle             *string      `json:"custom_title"`
		Delta                   string       `json:"delta"`
		Editors                 []string     `json:"editors"`
		ID                      string       `json:"id"`
		IgnoreDays              []string     `json:"ignore_days"`
		IgnoreZeroDays          bool         `json:"ignore_zero_days"`
		ImproveByPercent        *float64     `json:"improve_by_percent"`
		IsCurrentUserOwner      bool         `json:"is_current_user_owner"`
		IsEnabled               bool         `json:"is_enabled"`
		IsInverse               bool         `json:"is_inverse"`
		IsSnoozed               bool         `json:"is_snoozed"`
		IsTweeting              bool         `json:"is_tweeting"`
		Languages               []string     `json:"languages"`
		ModifiedAt              *string      `json:"modified_at"`
		Owner                   Owner        `json:"owner"`
		Projects                []string     `json:"projects"`
		RangeText               string       `json:"range_text"`
		Seconds                 int          `json:"seconds"`
		SharedWith              []string     `json:"shared_with"`
		SnoozeUntil             *string      `json:"snooze_until"`
		Status                  string       `json:"status"`
		StatusPercentCalculated int          `json:"status_percent_calculated"`
		Subscribers             []Subscriber `json:"subscribers"`
		Title                   string       `json:"title"`
		Type                    string       `json:"type"`
	}
)

// RenderToday generates a text representation from goal of the current day.
// If out is set to output.RawJSONOutput or output.JSONOutput, the goal will be marshaled to JSON.
// Expects exactly one summary for the current day. Will return an error otherwise.
func RenderToday(goal *Goal, out output.Output) (string, error) {
	if goal == nil {
		return "", errors.New("no goal found for the current day")
	}

	if out == output.RawJSONOutput {
		data, err := json.Marshal(goal)
		if err != nil {
			return "", fmt.Errorf("failed to marshal json goal: %s", err)
		}

		return string(data), nil
	}

	return goal.Data.ChartData[len(goal.Data.ChartData)-1].ActualSecondsText, nil
}
