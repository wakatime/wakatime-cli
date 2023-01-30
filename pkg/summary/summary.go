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

	// Range represents the the time range of a summary.
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

// RenderToday generates a text representation from summary of the current day.
// If out is set to output.RawJSONOutput or output.JSONOutput, the summary will be marshaled to JSON.
// Expects exactly one summary for the current day. Will return an error otherwise.
func RenderToday(summary *Summary, hideCategories bool, out output.Output) (string, error) {
	if summary == nil {
		return "", errors.New("no summary found for the current day")
	}

	if out == output.RawJSONOutput {
		data, err := json.Marshal(summary)
		if err != nil {
			return "", fmt.Errorf("failed to marshal json summary: %s", err)
		}

		return string(data), nil
	}

	if out == output.JSONOutput {
		type simplified struct {
			Text            string `json:"text"`
			HasTeamFeatures bool   `json:"has_team_features"`
		}

		s := simplified{
			Text:            summary.Data.GrandTotal.Text,
			HasTeamFeatures: summary.HasTeamFeatures,
		}

		data, err := json.Marshal(s)
		if err != nil {
			return "", fmt.Errorf("failed to marshal json simplified summary: %s", err)
		}

		return string(data), nil
	}

	if len(summary.Data.Categories) < 2 || hideCategories {
		return summary.Data.GrandTotal.Text, nil
	}

	var outputs []string
	for _, category := range summary.Data.Categories {
		outputs = append(outputs, fmt.Sprintf("%s %s", category.Text, category.Name))
	}

	return strings.Join(outputs, ", "), nil
}
