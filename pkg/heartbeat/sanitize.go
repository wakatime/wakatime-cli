package heartbeat

import (
	"path/filepath"
	"regexp"
)

// SanitizeConfig defines how a heartbeat should be sanitized.
type SanitizeConfig struct {
	HideBranchNames  []*regexp.Regexp
	HideFileNames    []*regexp.Regexp
	HideProjectNames []*regexp.Regexp
}

// WithSanitization initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline.
func WithSanitization(config SanitizeConfig) HandleOption {
	return func(next Handle) Handle {
		return func(hh []Heartbeat) ([]Result, error) {
			for n, h := range hh {
				hh[n] = Sanitize(h, config)
			}

			return next(hh)
		}
	}
}

// Sanitize accepts a heartbeat sanitizes it's sensitive data following passed
// in configuration and returns the sanitized version. On empty config will do nothing.
func Sanitize(h Heartbeat, config SanitizeConfig) Heartbeat {
	if len(h.Dependencies) == 0 {
		h.Dependencies = nil
	}

	if h.EntityType != FileType {
		return h
	}

	switch {
	case shouldSanitize(h.Entity, config.HideFileNames):
		h.Entity = "HIDDEN" + filepath.Ext(h.Entity)
		h = santizeMetaData(h)

		if h.Branch != nil && (len(config.HideBranchNames) == 0 || shouldSanitize(*h.Branch, config.HideBranchNames)) {
			h.Branch = nil
		}
	case h.Project != nil && shouldSanitize(*h.Project, config.HideProjectNames):
		h = santizeMetaData(h)
		if h.Branch != nil && (len(config.HideBranchNames) == 0 || shouldSanitize(*h.Branch, config.HideBranchNames)) {
			h.Branch = nil
		}
	case h.Branch != nil && shouldSanitize(*h.Branch, config.HideBranchNames):
		h.Branch = nil
	}

	return h
}

// santizeMetaData sanitizes metadata (curser position, dependencies, line number and lines).
func santizeMetaData(h Heartbeat) Heartbeat {
	h.CursorPosition = nil
	h.Dependencies = nil
	h.LineNumber = nil
	h.Lines = nil

	return h
}

// shouldSanitize checks a subject (entity, project, branch) of a heartbeat and
// checks it against the passed in regex patterns to determine, if this heartbeat
// should be sanitized.
func shouldSanitize(subject string, patterns []*regexp.Regexp) bool {
	for _, p := range patterns {
		if p.MatchString(subject) {
			return true
		}
	}

	return false
}
