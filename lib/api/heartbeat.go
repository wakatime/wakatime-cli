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
	Time           int64               `json:"time"`
	UserAgent      string              `json:"user_agent"`
}
