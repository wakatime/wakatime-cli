package api

import (
	"github.com/alanhamlett/wakatime-cli/lib/heartbeat/subtypes"
)

// Heartbeat is a heartbeat representation for sending to the wakatime api
type Heartbeat struct {
	Branch         *string             `json:"branch"`
	Category       subtypes.Category   `json:"category"`
	CursorPosition *int                `json:"cursorpos"`
	Dependencies   []string            `json:"dependencies"`
	Entity         string              `json:"entity"`
	EntityType     subtypes.EntityType `json:"type"`
	IsWrite        bool                `json:"is_write"`
	Language       string              `json:"language"`
	LineNumber     *int                `json:"lineno"`
	Lines          *int                `json:"lines"`
	Project        string              `json:"project"`
	Time           float64             `json:"time"`
	UserAgent      string              `json:"user_agent"`
}

// Int returns a pointer to the int value passed in.
func Int(v int) *int {
	return &v
}

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}
