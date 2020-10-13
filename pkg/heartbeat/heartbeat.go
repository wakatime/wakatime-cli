package heartbeat

import (
	"fmt"
	"runtime"

	"github.com/wakatime/wakatime-cli/pkg/version"

	"github.com/matishsiao/goInfo"
)

// Heartbeat is a structure representing activity for a user on a some entity.
type Heartbeat struct {
	Branch         *string    `json:"branch"`
	Category       Category   `json:"category"`
	CursorPosition *int       `json:"cursorpos"`
	Dependencies   []string   `json:"dependencies"`
	Entity         string     `json:"entity"`
	EntityType     EntityType `json:"type"`
	IsWrite        *bool      `json:"is_write"`
	Language       Language   `json:"language"`
	LineNumber     *int       `json:"lineno"`
	Lines          *int       `json:"lines"`
	Project        *string    `json:"project"`
	Time           float64    `json:"time"`
	UserAgent      string     `json:"user_agent"`
}

// ID returns an ID generated from the heartbeat data.
func (h Heartbeat) ID() string {
	var branch string
	if h.Branch != nil {
		branch = *h.Branch
	}

	var project string
	if h.Project != nil {
		project = *h.Project
	}

	var isWrite bool
	if h.IsWrite != nil {
		isWrite = *h.IsWrite
	}

	return fmt.Sprintf("%f-%s-%s-%s-%s-%s-%t",
		h.Time,
		h.EntityType,
		h.Category,
		project,
		branch,
		h.Entity,
		isWrite,
	)
}

// Result represents a response from the wakatime api.
type Result struct {
	Errors    []string
	Status    int
	Heartbeat Heartbeat
}

// Sender sends heartbeats to the wakatime api.
type Sender interface {
	Send(hh []Heartbeat) ([]Result, error)
}

// Handle does processing of heartbeats.
type Handle func(hh []Heartbeat) ([]Result, error)

// HandleOption is a function, which allows chaining multiple Handles.
type HandleOption func(next Handle) Handle

// NewHandle creates a new Handle, which acts like a processing pipeline,
// with a sender eventually sending the heartbeats.
func NewHandle(sender Sender, opts ...HandleOption) Handle {
	return func(hh []Heartbeat) ([]Result, error) {
		var h Handle = sender.Send
		for i := len(opts) - 1; i >= 0; i-- {
			h = opts[i](h)
		}

		return h(hh)
	}
}

// UserAgentUnknownPlugin generates a user agent from various system infos, including
// a default value for plugin.
func UserAgentUnknownPlugin() string {
	return UserAgent("Unknown/0")
}

// UserAgent generates a user agent from various system infos, including a
// a passed in value for plugin.
func UserAgent(plugin string) string {
	info := goInfo.GetInfo()

	return fmt.Sprintf(
		"wakatime/%s (%s-%s-%s) %s %s",
		version.Version,
		runtime.GOOS,
		info.Core,
		info.Platform,
		runtime.Version(),
		plugin,
	)
}

// Bool returns a pointer to the bool value passed in.
func Bool(v bool) *bool {
	return &v
}

// Int returns a pointer to the int value passed in.
func Int(v int) *int {
	return &v
}

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}
