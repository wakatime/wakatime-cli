package summary

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/output"
)

type (
	// Category represents the tracked category for a single day activity.
	Category struct {
		Decimal      string  `json:"decimal"`
		Digital      string  `json:"digital"`
		Hours        int     `json:"hours"`
		Minutes      int     `json:"minutes"`
		Name         string  `json:"name"`
		Percent      float64 `json:"percent"`
		Seconds      int     `json:"seconds"`
		Text         string  `json:"text"`
		TotalSeconds float64 `json:"total_seconds"`
	}

	// Counter represents the time counters.
	Counter struct {
		Decimal      string  `json:"decimal"`
		Digital      string  `json:"digital"`
		Hours        int     `json:"hours"`
		Minutes      int     `json:"minutes"`
		Name         string  `json:"name"`
		Percent      float64 `json:"percent"`
		Seconds      int     `json:"seconds"`
		Text         string  `json:"text"`
		TotalSeconds float64 `json:"total_seconds"`
	}

	// Data aggregates all activities for a single day.
	Data struct {
		Categories       []Category        `json:"categories"`
		Dependencies     []Dependency      `json:"dependencies"`
		Editors          []Editor          `json:"editors"`
		GrandTotal       GrandTotal        `json:"grand_total"`
		Languages        []Language        `json:"languages"`
		Machines         []Machine         `json:"machines"`
		OperatingSystems []OperatingSystem `json:"operating_systems"`
		Projects         []Project         `json:"projects"`
		Range            Range             `json:"range"`
	}

	// Dependency represents the discovered dependency for a single day activity.
	Dependency struct {
		Decimal      string  `json:"decimal"`
		Digital      string  `json:"digital"`
		Hours        int     `json:"hours"`
		Minutes      int     `json:"minutes"`
		Name         string  `json:"name"`
		Percent      float64 `json:"percent"`
		Seconds      int     `json:"seconds"`
		Text         string  `json:"text"`
		TotalSeconds float64 `json:"total_seconds"`
	}

	// Editor represents the used editor for a single day activity.
	Editor struct {
		Decimal      string  `json:"decimal"`
		Digital      string  `json:"digital"`
		Hours        int     `json:"hours"`
		Minutes      int     `json:"minutes"`
		Name         string  `json:"name"`
		Percent      float64 `json:"percent"`
		Seconds      int     `json:"seconds"`
		Text         string  `json:"text"`
		TotalSeconds float64 `json:"total_seconds"`
	}

	// GrandTotal represents the total working time for a single day.
	GrandTotal struct {
		Decimal      string  `json:"decimal"`
		Digital      string  `json:"digital"`
		Hours        int     `json:"hours"`
		Minutes      int     `json:"minutes"`
		Text         string  `json:"text"`
		TotalSeconds float64 `json:"total_seconds"`
	}

	// Language represents the used programming language for a single day activity.
	Language struct {
		Decimal      string  `json:"decimal"`
		Digital      string  `json:"digital"`
		Hours        int     `json:"hours"`
		Minutes      int     `json:"minutes"`
		Name         string  `json:"name"`
		Percent      float64 `json:"percent"`
		Seconds      int     `json:"seconds"`
		Text         string  `json:"text"`
		TotalSeconds float64 `json:"total_seconds"`
	}

	// Machine represents the used machine for a single day activity.
	Machine struct {
		Decimal       string  `json:"decimal"`
		Digital       string  `json:"digital"`
		Hours         int     `json:"hours"`
		MachineNameID string  `json:"machine_name_id"`
		Minutes       int     `json:"minutes"`
		Name          string  `json:"name"`
		Percent       float64 `json:"percent"`
		Seconds       int     `json:"seconds"`
		Text          string  `json:"text"`
		TotalSeconds  float64 `json:"total_seconds"`
	}

	// OperatingSystem represents the used operating system for a single day activity.
	OperatingSystem struct {
		Decimal      string  `json:"decimal"`
		Digital      string  `json:"digital"`
		Hours        int     `json:"hours"`
		Minutes      int     `json:"minutes"`
		Name         string  `json:"name"`
		Percent      float64 `json:"percent"`
		Seconds      int     `json:"seconds"`
		Text         string  `json:"text"`
		TotalSeconds float64 `json:"total_seconds"`
	}

	// Project represents the discovered project for a single day activity.
	Project struct {
		Decimal      string  `json:"decimal"`
		Digital      string  `json:"digital"`
		Hours        int     `json:"hours"`
		Minutes      int     `json:"minutes"`
		Name         string  `json:"name"`
		Percent      float64 `json:"percent"`
		Seconds      int     `json:"seconds"`
		Text         string  `json:"text"`
		TotalSeconds float64 `json:"total_seconds"`
	}

	// Range represents the time range of a summary.
	Range struct {
		Date     string `json:"date"`
		End      string `json:"end"`
		Start    string `json:"start"`
		Text     string `json:"text"`
		Timezone string `json:"timezone"`
	}

	// Summary represents the tracked working time for a single day.
	Summary struct {
		CachedAt        string `json:"cached_at"`
		Data            Data   `json:"data"`
		HasTeamFeatures bool   `json:"has_team_features"`
	}
)

// TimeFormat represents the duration format of the status bar text.
type TimeFormat int

const (
	// TextTimeFormat formats durations like "11 hrs 41 mins". This is the default value.
	TextTimeFormat TimeFormat = iota
	// ShortTimeFormat formats durations like "11h 41m".
	ShortTimeFormat
	// DecimalTimeFormat formats durations like "11.7h".
	DecimalTimeFormat
	// DigitalTimeFormat formats durations like "11:41".
	DigitalTimeFormat
)

// ParseTimeFormat parses a time format from a string.
func ParseTimeFormat(s string) (TimeFormat, error) {
	switch s {
	case "text":
		return TextTimeFormat, nil
	case "short":
		return ShortTimeFormat, nil
	case "decimal":
		return DecimalTimeFormat, nil
	case "digital":
		return DigitalTimeFormat, nil
	default:
		return TextTimeFormat, fmt.Errorf("invalid time format %q", s)
	}
}

type duration struct {
	hours        int
	minutes      int
	text         string
	totalSeconds float64
}

// RenderToday generates a text representation from summary of the current day.
// If out is set to output.RawJSONOutput or output.JSONOutput, the summary will be marshaled to JSON.
// Expects exactly one summary for the current day. Will return an error otherwise.
func RenderToday(
	summary *Summary,
	hideCategories bool,
	hideMinutes bool,
	timeFormat TimeFormat,
	maxCategories int,
	out output.Output,
) (string, error) {
	if summary == nil {
		return renderEmptyToday(out, hideMinutes, timeFormat), errors.New("no summary found for the current day")
	}

	if out == output.RawJSONOutput {
		data, err := json.Marshal(summary)
		if err != nil {
			return renderEmptyToday(out, hideMinutes, timeFormat), fmt.Errorf("failed to marshal json summary: %s", err)
		}

		return string(data), nil
	}

	if out == output.JSONOutput {
		type simplified struct {
			Text            string `json:"text"`
			HasTeamFeatures bool   `json:"has_team_features"`
		}

		s := simplified{
			Text:            getText(summary, hideCategories, hideMinutes, timeFormat, maxCategories),
			HasTeamFeatures: summary.HasTeamFeatures,
		}

		data, err := json.Marshal(s)
		if err != nil {
			return renderEmptyToday(out, hideMinutes, timeFormat),
				fmt.Errorf("failed to marshal json simplified summary: %s", err)
		}

		return string(data), nil
	}

	return getText(summary, hideCategories, hideMinutes, timeFormat, maxCategories), nil
}

func getText(
	summary *Summary,
	hideCategories bool,
	hideMinutes bool,
	timeFormat TimeFormat,
	maxCategories int,
) string {
	if len(summary.Data.Categories) < 2 || hideCategories {
		total := summary.Data.GrandTotal

		return durationText(duration{
			hours:        total.Hours,
			minutes:      total.Minutes,
			text:         total.Text,
			totalSeconds: total.TotalSeconds,
		}, hideMinutes, timeFormat)
	}

	var outputs []string

	categories := summary.Data.Categories

	if maxCategories > 0 && len(categories) > maxCategories {
		categories = categories[:maxCategories]
	}

	for _, category := range categories {
		text := durationText(duration{
			hours:        category.Hours,
			minutes:      category.Minutes,
			text:         category.Text,
			totalSeconds: category.TotalSeconds,
		}, hideMinutes, timeFormat)
		outputs = append(outputs, fmt.Sprintf("%s %s", text, category.Name))
	}

	result := strings.Join(outputs, ", ")

	if maxCategories > 1 && len(summary.Data.Categories) > maxCategories {
		result += "..."
	}

	return result
}

func durationText(d duration, hideMinutes bool, timeFormat TimeFormat) string {
	switch timeFormat {
	case ShortTimeFormat:
		return shortDurationText(d, hideMinutes)
	case DecimalTimeFormat:
		return fmt.Sprintf("%.1fh", d.totalSeconds/3600)
	case DigitalTimeFormat:
		return fmt.Sprintf("%d:%02d", d.hours, d.minutes)
	case TextTimeFormat:
		return textDurationText(d, hideMinutes)
	default:
		return d.text
	}
}

func textDurationText(d duration, hideMinutes bool) string {
	if !hideMinutes || d.hours == 0 {
		return d.text
	}

	if d.hours == 1 {
		return "1 hr"
	}

	return fmt.Sprintf("%d hrs", d.hours)
}

func shortDurationText(d duration, hideMinutes bool) string {
	switch {
	case d.hours > 0 && hideMinutes:
		return fmt.Sprintf("%dh", d.hours)
	case d.hours > 0:
		return fmt.Sprintf("%dh %dm", d.hours, d.minutes)
	case d.minutes > 0:
		return fmt.Sprintf("%dm", d.minutes)
	default:
		return fmt.Sprintf("%ds", int(d.totalSeconds))
	}
}

func renderEmptyToday(out output.Output, hideMinutes bool, timeFormat TimeFormat) string {
	text := "0 secs"
	if hideMinutes {
		text = "0 hrs"
	}

	if timeFormat != TextTimeFormat {
		text = durationText(duration{}, hideMinutes, timeFormat)
	}

	if out == output.TextOutput {
		return text
	}

	if out == output.RawJSONOutput {
		return fmt.Sprintf(`{"data":{"categories":[],"dependencies":[],"editors":[],`+
			`"grand_total":{"decimal":"0.00","digital":"0:00","hours":0,"minutes":0,"text":"%s",`+
			`"total_seconds":0},"languages":[],"machines":[],"operating_systems":[],"projects":[]}}`, text)
	}

	return fmt.Sprintf(`{"text":"%s"}`, text)
}
