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
	// BranchPatterns will be matched against the entity file path and if matching, will obfuscate it.
	BranchPatterns []regex.Regex
	// DependencyPatterns will be matched against the entity file path and if matching, will omit all dependencies.
	DependencyPatterns []regex.Regex
	// FilePatterns will be matched against a file entity's name and if matching will obfuscate
	// the file name and common heartbeat meta data (cursor position, dependencies, line number and lines).
	FilePatterns []regex.Regex
	// HideProjectFolder determines if project folder should be obfuscated.
	HideProjectFolder bool
	// ProjectPatterns will be matched against the entity file path and if matching will obfuscate
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

	check := SanitizeCheck{
		Entity:              h.Entity,
		ProjectPath:         h.ProjectPath,
		ProjectPathOverride: h.ProjectPathOverride,
	}

	// project patterns
	if h.Project != nil {
		check.Patterns = config.ProjectPatterns
		if ShouldSanitize(ctx, check) {
			h = sanitizeMetaData(h)
		}
	}

	// file patterns
	check.Patterns = config.FilePatterns
	if ShouldSanitize(ctx, check) {
		if h.EntityType == FileType {
			h.Entity = "HIDDEN" + filepath.Ext(h.Entity)
		} else {
			h.Entity = "HIDDEN"
		}

		if len(config.BranchPatterns) == 0 {
			h.Branch = nil
		}

		if len(config.DependencyPatterns) == 0 {
			h.Dependencies = nil
		}

		h = sanitizeMetaData(h)
	}

	// branch patterns
	if h.Branch != nil {
		check.Patterns = config.BranchPatterns
		if ShouldSanitize(ctx, check) {
			h.Branch = nil
		}
	}

	// dependency patterns
	if h.Dependencies != nil {
		check.Patterns = config.DependencyPatterns
		if ShouldSanitize(ctx, check) {
			h.Dependencies = nil
		}
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

// sanitizeMetaData sanitizes metadata (cursor position, line number, lines and project root count).
func sanitizeMetaData(h Heartbeat) Heartbeat {
	h.CursorPosition = nil
	h.LineNumber = nil
	h.Lines = nil
	h.ProjectRootCount = nil

	return h
}

// SanitizeCheck defines a configuration for checking if a heartbeat should be sanitized.
type SanitizeCheck struct {
	Entity              string
	Patterns            []regex.Regex
	ProjectPath         string
	ProjectPathOverride string
}

// ShouldSanitize checks the entity filepath or project path of a heartbeat
// against the passed in regex patterns to determine, if this heartbeat
// should be sanitized.
func ShouldSanitize(ctx context.Context, check SanitizeCheck) bool {
	for _, p := range check.Patterns {
		if p.MatchString(ctx, check.Entity) {
			return true
		}

		if p.MatchString(ctx, check.ProjectPath) {
			return true
		}

		if p.MatchString(ctx, check.ProjectPathOverride) {
			return true
		}
	}

	return false
}
