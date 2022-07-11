package heartbeat

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/system"
	"github.com/wakatime/wakatime-cli/pkg/version"

	"github.com/matishsiao/goInfo"
)

// remoteAddressRegex is a pattern for (ssh|sftp)://user:pass@host:port.
var remoteAddressRegex = regexp.MustCompile(`(?i)^((ssh|sftp)://)+(?P<credentials>[^:@]+(:([^:@])+)?@)?[^:]+(:\d+)?`)

// Heartbeat is a structure representing activity for a user on a some entity.
type Heartbeat struct {
	ApiKey                string     `json:"-"`
	Branch                *string    `json:"branch"`
	BranchAlternate       string     `json:"-"`
	Category              Category   `json:"category"`
	CursorPosition        *int       `json:"cursorpos"`
	Dependencies          []string   `json:"dependencies"`
	Entity                string     `json:"entity"`
	EntityType            EntityType `json:"type"`
	IsUnsavedEntity       bool       `json:"-"`
	IsWrite               *bool      `json:"is_write"`
	Language              *string    `json:"language"`
	LanguageAlternate     string     `json:"-"`
	LineNumber            *int       `json:"lineno"`
	Lines                 *int       `json:"lines"`
	LocalFile             string     `json:"-"`
	LocalFileNeedsCleanup bool       `json:"-"`
	Project               *string    `json:"project"`
	ProjectAlternate      string     `json:"-"`
	ProjectOverride       string     `json:"-"`
	ProjectPath           string     `json:"-"`
	ProjectPathOverride   string     `json:"-"`
	Time                  float64    `json:"time"`
	UserAgent             string     `json:"user_agent"`
}

// New creates a new instance of Heartbeat with formatted entity
// and local file paths for file type heartbeats.
func New(
	branchAlternate string,
	category Category,
	cursorPosition *int,
	entity string,
	entityType EntityType,
	isUnsavedEntity bool,
	isWrite *bool,
	language *string,
	languageAlternate string,
	lineNumber *int,
	lines *int,
	localFile string,
	projectAlternate string,
	projectOverride string,
	projectPathOverride string,
	time float64,
	userAgent string,
) Heartbeat {
	return Heartbeat{
		BranchAlternate:     branchAlternate,
		Category:            category,
		CursorPosition:      cursorPosition,
		Entity:              entity,
		EntityType:          entityType,
		IsUnsavedEntity:     isUnsavedEntity,
		IsWrite:             isWrite,
		Language:            language,
		LanguageAlternate:   languageAlternate,
		LineNumber:          lineNumber,
		Lines:               lines,
		LocalFile:           localFile,
		ProjectAlternate:    projectAlternate,
		ProjectOverride:     projectOverride,
		ProjectPathOverride: projectPathOverride,
		Time:                time,
		UserAgent:           userAgent,
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

// IsRemote returns true when entity is a remote file.
func (h Heartbeat) IsRemote() bool {
	if h.EntityType != FileType {
		return false
	}

	if h.IsUnsavedEntity {
		return false
	}

	return remoteAddressRegex.MatchString(h.Entity)
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

// Sender is a noop api client, used to always skip sending to the API.
type Noop struct{}

// SendHeartbeats always returns nil.
func (Noop) SendHeartbeats(_ []Heartbeat) ([]Result, error) {
	return nil, nil
}

// Handle does processing of heartbeats.
type Handle func(hh []Heartbeat) ([]Result, error)

// HandleOption is a function, which allows chaining multiple Handles.
type HandleOption func(next Handle) Handle

// NewHandle creates a new Handle, which acts like a processing pipeline,
// with a sender eventually sending the heartbeats.
func NewHandle(sender Sender, opts ...HandleOption) Handle {
	return func(heartbeats []Heartbeat) ([]Result, error) {
		var handle Handle = sender.SendHeartbeats
		for i := len(opts) - 1; i >= 0; i-- {
			handle = opts[i](handle)
		}

		return handle(heartbeats)
	}
}

// UserAgent generates a user agent from various system infos, including a
// a passed in value for plugin.
func UserAgent(plugin string) string {
	info, err := goInfo.GetInfo()
	if err != nil {
		log.Debugf("goInfo.GetInfo error: %s", err)
	}

	if plugin == "" {
		plugin = "Unknown/0"
	}

	return fmt.Sprintf(
		"wakatime/%s (%s-%s-%s) %s %s",
		version.Version,
		system.OSName(),
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

// PointerTo returns a pointer to the value passed in.
func PointerTo[t bool | int | string](v t) *t {
	return &v
}

func isDir(filepath string) bool {
	info, _ := os.Stat(filepath)
	return info.IsDir()
}
