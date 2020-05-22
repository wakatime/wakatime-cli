package summary

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Category represents tracked working time of a specific category.
type Category struct {
	Category string
	Total    string
}

// Summary represents the tracked working time for a single day.
type Summary struct {
	Date       time.Time
	Total      string
	ByCategory []Category
}

// RenderToday generates a text representation from summaries of the current day.
// Expects exactly one summary for the current day. Will return an error otherwise.
func RenderToday(summaries []Summary) (string, error) {
	var (
		now   = time.Now()
		today []Summary
	)

	for _, s := range summaries {
		if now.Year() != s.Date.Year() || now.Month() != s.Date.Month() || now.Day() != s.Date.Day() {
			continue
		}

		if len(today) > 0 {
			return "", fmt.Errorf("received two summaries for the current day. 1. %+v, 2. %+v", today[0], s)
		}

		today = append(today, s)
	}

	if len(today) == 0 {
		return "", errors.New("no summary found for the current day")
	}

	if len(today[0].ByCategory) < 2 {
		return string(today[0].Total), nil
	}

	var outputs []string
	for _, category := range today[0].ByCategory {
		outputs = append(outputs, fmt.Sprintf("%s %s", category.Total, category.Category))
	}

	return strings.Join(outputs, ", "), nil
}
