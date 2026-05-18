package heartbeat

import (
	"context"
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
	AILineChanges         *int       `json:"ai_line_changes,omitempty"`
	AISession             string     `json:"ai_session,omitempty"`
	AISubscriptionPlan    string     `json:"ai_subscription_plan,omitempty"`
	AIInputTokens         int64      `json:"ai_input_tokens,omitempty"`
	AIOutputTokens        int64      `json:"ai_output_tokens,omitempty"`
	AIPromptLength        int        `json:"ai_prompt_length,omitempty"`
	APIKey                string     `json:"-"`
	APIURL                string     `json:"-"`
	Branch                *string    `json:"branch,omitempty"`
	BranchAlternate       string     `json:"-"`
	Category              string     `json:"category,omitempty"`
	CursorPosition        *int       `json:"cursorpos,omitempty"`
	Dependencies          []string   `json:"dependencies,omitempty"`
	Entity                string     `json:"entity"`
	EntityType            EntityType `json:"type"`
	HumanLineChanges      *int       `json:"human_line_changes,omitempty"`
	IsUnsavedEntity       bool       `json:"-"`
	IsWrite               *bool      `json:"is_write,omitempty"`
	Language              *string    `json:"language,omitempty"`
	LanguageAlternate     string     `json:"-"`
	LineNumber            *int       `json:"lineno,omitempty"`
	Lines                 *int       `json:"lines,omitempty"`
	LocalFile             string     `json:"-"`
	LocalFileNeedsCleanup bool       `json:"-"`
	Project               *string    `json:"project,omitempty"`
	ProjectAlternate      string     `json:"-"`
	ProjectFromGitRemote  bool       `json:"-"`
	ProjectOverride       string     `json:"-"`
	ProjectPath           string     `json:"-"`
	ProjectPathOverride   string     `json:"-"`
	ProjectRootCount      *int       `json:"project_root_count,omitempty"`
	Time                  float64    `json:"time"`
	UserAgent             string     `json:"user_agent"`
}

// AITokens contains the previous and current token counts for calculating the delta input and output AI tokens used
// since the last heartbeat.
type AITokens struct {
	LastInput     int64
	LastOutput    int64
	CurrentInput  int64
	CurrentOutput int64
}

// New creates a new instance of Heartbeat with formatted entity
// and local file paths for file type heartbeats.
func New(
	aiLineChanges *int,
	branchAlternate string,
	category string,
	cursorPosition *int,
	entity string,
	entityType EntityType,
	humanLineChanges *int,
	isUnsavedEntity bool,
	isWrite *bool,
	language *string,
	languageAlternate string,
	lineNumber *int,
	lines *int,
	localFile string,
	projectAlternate string,
	projectFromGitRemote bool,
	projectOverride string,
	projectPathOverride string,
	time float64,
	userAgent string,
) Heartbeat {
	return Heartbeat{
		AILineChanges:        aiLineChanges,
		BranchAlternate:      branchAlternate,
		Category:             category,
		CursorPosition:       cursorPosition,
		Entity:               entity,
		EntityType:           entityType,
		HumanLineChanges:     humanLineChanges,
		IsUnsavedEntity:      isUnsavedEntity,
		IsWrite:              isWrite,
		Language:             language,
		LanguageAlternate:    languageAlternate,
		LineNumber:           lineNumber,
		Lines:                lines,
		LocalFile:            localFile,
		ProjectAlternate:     projectAlternate,
		ProjectFromGitRemote: projectFromGitRemote,
		ProjectOverride:      projectOverride,
		ProjectPathOverride:  projectPathOverride,
		Time:                 time,
		UserAgent:            userAgent,
	}
}

// NewWithAITokens creates a new instance of Heartbeat with formatted entity
// and local file paths for file type heartbeats.
func NewWithAITokens(
	aiLineChanges *int,
	aiSession string,
	aiTokens AITokens,
	branchAlternate string,
	category string,
	cursorPosition *int,
	entity string,
	entityType EntityType,
	humanLineChanges *int,
	isUnsavedEntity bool,
	isWrite *bool,
	language *string,
	languageAlternate string,
	lineNumber *int,
	lines *int,
	localFile string,
	projectAlternate string,
	projectFromGitRemote bool,
	projectOverride string,
	projectPathOverride string,
	time float64,
	userAgent string,
) Heartbeat {
	inputTokens := aiTokens.CurrentInput - aiTokens.LastInput
	if inputTokens < 0 {
		inputTokens = 0
	}

	outputTokens := aiTokens.CurrentOutput - aiTokens.LastOutput
	if outputTokens < 0 {
		outputTokens = 0
	}

	return Heartbeat{
		AILineChanges:        aiLineChanges,
		AISession:            aiSession,
		AIInputTokens:        inputTokens,
		AIOutputTokens:       outputTokens,
		BranchAlternate:      branchAlternate,
		Category:             category,
		CursorPosition:       cursorPosition,
		Entity:               entity,
		EntityType:           entityType,
		HumanLineChanges:     humanLineChanges,
		IsUnsavedEntity:      isUnsavedEntity,
		IsWrite:              isWrite,
		Language:             language,
		LanguageAlternate:    languageAlternate,
		LineNumber:           lineNumber,
		Lines:                lines,
		LocalFile:            localFile,
		ProjectAlternate:     projectAlternate,
		ProjectFromGitRemote: projectFromGitRemote,
		ProjectOverride:      projectOverride,
		ProjectPathOverride:  projectPathOverride,
		Time:                 time,
		UserAgent:            userAgent,
	}
}

// ID returns an ID generated from the heartbeat data.
func (h Heartbeat) ID() string {
	branch := "unset"
	if h.Branch != nil {
		branch = *h.Branch
	}

	project := "unset"
	if h.Project != nil {
		project = *h.Project
	}

	var isWrite bool
	if h.IsWrite != nil {
		isWrite = *h.IsWrite
	}

	cursorPos := "nil"
	if h.CursorPosition != nil {
		cursorPos = fmt.Sprint(*h.CursorPosition)
	}

	category := "undefined"
	if h.Category != "" {
		category = h.Category
	}

	return fmt.Sprintf("%f-%s-%s-%s-%s-%s-%s-%t",
		h.Time,
		cursorPos,
		h.EntityType,
		category,
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
	Errors []string
	Status int
	ID     string
	// Heartbeat is the original heartbeat that was sent.
	Heartbeat Heartbeat
	// it's a temporary solution before we have a better way to handle (avoid import cycle)
	FileExpert any
}

// Sender sends heartbeats to the wakatime api.
type Sender interface {
	SendHeartbeats(context.Context, []Heartbeat) ([]Result, error)
}

// Handle does processing of heartbeats.
type Handle func(context.Context, []Heartbeat) ([]Result, error)

// HandleOption is a function, which allows chaining multiple Handles.
type HandleOption func(next Handle) Handle

// NewHandle creates a new Handle, which acts like a processing pipeline,
// with a sender eventually sending the heartbeats.
func NewHandle(sender Sender, opts ...HandleOption) Handle {
	return func(ctx context.Context, hh []Heartbeat) ([]Result, error) {
		var handle Handle = sender.SendHeartbeats
		for i := len(opts) - 1; i >= 0; i-- {
			handle = opts[i](handle)
		}

		return handle(ctx, hh)
	}
}

// UserAgent generates a user agent from various system infos, including a
// a passed in value for plugin.
func UserAgent(ctx context.Context, plugin string) (userAgent string) {
	logger := log.Extract(ctx)
	template := "wakatime/%s (%s-%s-%s) %s %s"

	defer func() {
		if r := recover(); r != nil {
			userAgent = fmt.Sprintf(
				template,
				version.Version,
				strings.TrimSpace(system.OSName(ctx)),
				"unknown",
				"unknown",
				strings.TrimSpace(runtime.Version()),
				strings.TrimSpace(plugin),
			)
		}
	}()

	if plugin == "" {
		plugin = "Unknown/0"
	}

	plugin = normalizePluginVersions(plugin)

	info, err := goInfo.GetInfo()
	if err != nil {
		logger.Debugf("goInfo.GetInfo error: %s", err)
	}

	userAgent = fmt.Sprintf(
		template,
		version.Version,
		strings.TrimSpace(system.OSName(ctx)),
		strings.TrimSpace(info.Core),
		strings.TrimSpace(info.Platform),
		strings.TrimSpace(runtime.Version()),
		strings.TrimSpace(plugin),
	)

	return userAgent
}

// normalizePluginVersions replaces empty product versions with "unknown".
// Product tokens in user agents are space-separated, so `Claude/` becomes
// `Claude/unknown` while already-populated tokens are left unchanged.
func normalizePluginVersions(plugin string) string {
	fields := strings.Fields(plugin)
	changed := false

	for i, field := range fields {
		if strings.HasSuffix(field, "/") {
			fields[i] = field + "unknown"
			changed = true
		}
	}

	if !changed {
		return strings.TrimSpace(plugin)
	}

	return strings.Join(fields, " ")
}

// PointerTo returns a pointer to the value passed in.
func PointerTo[t bool | int | string](v t) *t {
	return &v
}

func isDir(ctx context.Context, filepath string) bool {
	logger := log.Extract(ctx)

	info, err := os.Stat(filepath)
	if err != nil {
		logger.Warnf("failed to stat filepath %q: %s", filepath, err)
		return false
	}

	return info.IsDir()
}
