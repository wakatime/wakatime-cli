package heartbeat

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/version"
	"github.com/wakatime/wakatime-cli/pkg/windows"

	"github.com/matishsiao/goInfo"
	"github.com/yookoala/realpath"
)

// Heartbeat is a structure representing activity for a user on a some entity.
type Heartbeat struct {
	Branch            *string    `json:"branch"`
	Category          Category   `json:"category"`
	CursorPosition    *int       `json:"cursorpos"`
	Dependencies      []string   `json:"dependencies"`
	Entity            string     `json:"entity"`
	EntityType        EntityType `json:"type"`
	IsWrite           *bool      `json:"is_write"`
	Language          *string    `json:"language"`
	LanguageAlternate string     `json:"-"`
	LineNumber        *int       `json:"lineno"`
	Lines             *int       `json:"lines"`
	LocalFile         string     `json:"-"`
	Project           *string    `json:"project"`
	ProjectAlternate  string     `json:"-"`
	ProjectOverride   string     `json:"-"`
	Time              float64    `json:"time"`
	UserAgent         string     `json:"user_agent"`
}

// New creates a new instance of Heartbeat with formatted entity
// and local file paths for file type heartbeats.
func New(
	category Category,
	cursorPosition *int,
	entity string,
	entityType EntityType,
	isWrite *bool,
	language *string,
	languageAlternate string,
	lineNumber *int,
	localFile string,
	projectAlternate string,
	projectOverride string,
	time float64,
	userAgent string,
) Heartbeat {
	if entityType == FileType {
		formatted, err := filepath.Abs(entity)
		if err != nil {
			log.Warnf("failed to resolve the absolute path of %q: %s", entity, err)
		} else {
			entity = formatted
		}

		formatted, err = realpath.Realpath(entity)
		if err != nil {
			log.Warnf("failed to resolve the real path of %q: %s", entity, err)
		} else {
			entity = formatted
		}
	}

	if entityType == FileType && runtime.GOOS == "windows" {
		formatted, err := windows.FormatFilePath(entity)
		if err != nil {
			log.Warnf("failed to format windows file path: %q: %s", entity, err)
		} else {
			entity = formatted
		}

		localFile, err = windows.FormatLocalFilePath(localFile, entity)
		if err != nil {
			log.Warnf("failed to format local file path: %s", err)
		}
	}

	return Heartbeat{
		Category:          category,
		CursorPosition:    cursorPosition,
		Entity:            entity,
		EntityType:        entityType,
		IsWrite:           isWrite,
		Language:          language,
		LanguageAlternate: languageAlternate,
		LineNumber:        lineNumber,
		LocalFile:         localFile,
		ProjectAlternate:  projectAlternate,
		ProjectOverride:   projectOverride,
		Time:              time,
		UserAgent:         userAgent,
	}
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
	SendHeartbeats(hh []Heartbeat) ([]Result, error)
}

// Handle does processing of heartbeats.
type Handle func(hh []Heartbeat) ([]Result, error)

// HandleOption is a function, which allows chaining multiple Handles.
type HandleOption func(next Handle) Handle

// NewHandle creates a new Handle, which acts like a processing pipeline,
// with a sender eventually sending the heartbeats.
func NewHandle(sender Sender, opts ...HandleOption) Handle {
	return func(hh []Heartbeat) ([]Result, error) {
		var h Handle = sender.SendHeartbeats
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

// PluginFromUserAgent parses the plugin name from a wakatime user agent.
func PluginFromUserAgent(userAgent string) string {
	splitted := strings.Split(userAgent, " ")
	splitted = strings.Split(splitted[len(splitted)-1], "/")
	splitted = strings.Split(splitted[0], "-")

	return splitted[0]
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
