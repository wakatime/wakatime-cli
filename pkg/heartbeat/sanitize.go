package heartbeat

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/regex"
)

// SanitizeConfig defines how a heartbeat should be sanitized.
type SanitizeConfig struct {
	// BranchPatterns will be matched against the branch and if matching, will obfuscate it.
	BranchPatterns []regex.Regex
	// FilePatterns will be matched against a file entity's name and if matching will obfuscate
	// the file name and common heartbeat meta data (cursor position, dependencies, line number and lines).
	FilePatterns []regex.Regex
	// HideProjectFolder determines if project folder should be obfuscated.
	HideProjectFolder bool
	// ProjectPatterns will be matched against the project name and if matching will obfuscate
	// common heartbeat meta data (cursor position, dependencies, line number and lines).
	ProjectPatterns []regex.Regex
}

// WithSanitization initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to hide sensitive data.
func WithSanitization(config SanitizeConfig) HandleOption {
	return func(next Handle) Handle {
		return func(ctx context.Context, hh []Heartbeat) ([]Result, error) {
			logger := log.Extract(ctx)
			logger.Debugln("execute heartbeat sanitization")

			for n, h := range hh {
				hh[n] = Sanitize(ctx, h, config)
			}

			return next(ctx, hh)
		}
	}
}

// Sanitize accepts a heartbeat sanitizes it's sensitive data following passed
// in configuration and returns the sanitized version. On empty config will do nothing.
func Sanitize(ctx context.Context, h Heartbeat, config SanitizeConfig) Heartbeat {
	if len(h.Dependencies) == 0 {
		h.Dependencies = nil
	}

	switch {
	case ShouldSanitize(ctx, h.Entity, config.FilePatterns):
		if h.EntityType == FileType {
			h.Entity = "HIDDEN" + filepath.Ext(h.Entity)
		} else {
			h.Entity = "HIDDEN"
		}

		h = sanitizeMetaData(h)

		if h.Branch != nil && (len(config.BranchPatterns) == 0 || ShouldSanitize(ctx, *h.Branch, config.BranchPatterns)) {
			h.Branch = nil
		}
	case h.Project != nil && ShouldSanitize(ctx, *h.Project, config.ProjectPatterns):
		h = sanitizeMetaData(h)
		if h.Branch != nil && (len(config.BranchPatterns) == 0 || ShouldSanitize(ctx, *h.Branch, config.BranchPatterns)) {
			h.Branch = nil
		}
	case h.Branch != nil && ShouldSanitize(ctx, *h.Branch, config.BranchPatterns):
		h.Branch = nil
	}

	h = hideProjectFolder(h, config.HideProjectFolder)

	h = hideCredentials(h)

	return h
}

// hideProjectFolder makes entity relative to project folder if we're hiding the project folder.
func hideProjectFolder(h Heartbeat, hideProjectFolder bool) Heartbeat {
	if h.EntityType != FileType || !hideProjectFolder {
		return h
	}

	if h.ProjectPath != "" {
		// this makes entity path relative after trim
		if !strings.HasSuffix(h.ProjectPath, "/") {
			h.ProjectPath += "/"
		}

		if strings.HasPrefix(h.Entity, h.ProjectPath) {
			h.Entity = strings.TrimPrefix(h.Entity, h.ProjectPath)
			h.ProjectRootCount = nil

			return h
		}
	}

	if h.ProjectPathOverride != "" {
		// this makes entity path relative after trim
		if !strings.HasSuffix(h.ProjectPathOverride, "/") {
			h.ProjectPathOverride += "/"
		}

		h.Entity = strings.TrimPrefix(h.Entity, h.ProjectPathOverride)
		h.ProjectRootCount = nil
	}

	return h
}

func hideCredentials(h Heartbeat) Heartbeat {
	if !h.IsRemote() {
		return h
	}

	match := remoteAddressRegex.FindStringSubmatch(h.Entity)
	paramsMap := make(map[string]string)

	for i, name := range remoteAddressRegex.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}

	if creds, ok := paramsMap["credentials"]; ok {
		h.Entity = strings.ReplaceAll(h.Entity, creds, "")
	}

	return h
}

// sanitizeMetaData sanitizes metadata (cursor position, dependencies, line number and lines).
func sanitizeMetaData(h Heartbeat) Heartbeat {
	h.CursorPosition = nil
	h.Dependencies = nil
	h.LineNumber = nil
	h.Lines = nil
	h.ProjectRootCount = nil

	return h
}

// ShouldSanitize checks a subject (entity, project, branch) of a heartbeat and
// checks it against the passed in regex patterns to determine, if this heartbeat
// should be sanitized.
func ShouldSanitize(ctx context.Context, subject string, patterns []regex.Regex) bool {
	for _, p := range patterns {
		if p.MatchString(ctx, subject) {
			return true
		}
	}

	return false
}
