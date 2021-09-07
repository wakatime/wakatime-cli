package summary

import (
	"errors"
	"fmt"
	"strings"
)

// Category represents tracked working time of a specific category.
type Category struct {
	Category string
	Total    string
}

// Summary represents the tracked working time for a single day.
type Summary struct {
	Total      string
	ByCategory []Category
}

// RenderToday generates a text representation from summary of the current day.
// Expects exactly one summary for the current day. Will return an error otherwise.
func RenderToday(summary *Summary, statusBarHideCategories bool) (string, error) {
	if summary == nil {
		return "", errors.New("no summary found for the current day")
	}

	if len(summary.ByCategory) < 2 || statusBarHideCategories {
		return string(summary.Total), nil
	}

	var outputs []string
	for _, category := range summary.ByCategory {
		outputs = append(outputs, fmt.Sprintf("%s %s", category.Total, category.Category))
	}

	return strings.Join(outputs, ", "), nil
}
