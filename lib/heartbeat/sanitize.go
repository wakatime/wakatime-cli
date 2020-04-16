package heartbeat

import (
	"path/filepath"
	"regexp"

	"github.com/alanhamlett/wakatime-cli/lib/api"
	"github.com/alanhamlett/wakatime-cli/lib/heartbeat/subtypes"
)

// Obfuscate defines how a heartbeat should be sanitized
type Obfuscate struct {
	HideBranchNames  []*regexp.Regexp
	HideFileNames    []*regexp.Regexp
	HideProjectNames []*regexp.Regexp
}

// Sanitize creates a new heartbeat request representation with optionally
// sanitizing sensitive data
func Sanitize(h Heartbeat, obfuscate Obfuscate) api.Heartbeat {
	dependencies := h.Dependencies
	if len(dependencies) == 0 {
		dependencies = nil
	}

	return sanitize(api.Heartbeat{
		Branch:         &h.Branch,
		Category:       h.Category,
		CursorPosition: &h.CursorPosition,
		Dependencies:   dependencies,
		Entity:         h.Entity,
		EntityType:     h.EntityType,
		IsWrite:        h.IsWrite,
		Language:       h.Language,
		LineNumber:     &h.LineNumber,
		Lines:          &h.Lines,
		Project:        h.Project,
		Time:           float64(h.Time),
		UserAgent:      h.UserAgent,
	}, obfuscate)
}

func sanitize(h api.Heartbeat, obfuscate Obfuscate) api.Heartbeat {
	if h.EntityType != subtypes.FileType {
		return h
	}

	if shouldObfuscate(h.Entity, obfuscate.HideFileNames) {
		h.Entity = "HIDDEN" + filepath.Ext(h.Entity)
		h = santizeMetaData(h)
		if len(obfuscate.HideBranchNames) == 0 || shouldObfuscate(*h.Branch, obfuscate.HideBranchNames) {
			h.Branch = nil
		}
	} else if shouldObfuscate(h.Project, obfuscate.HideProjectNames) {
		h = santizeMetaData(h)
		if len(obfuscate.HideBranchNames) == 0 || shouldObfuscate(*h.Branch, obfuscate.HideBranchNames) {
			h.Branch = nil
		}
	} else if shouldObfuscate(*h.Branch, obfuscate.HideBranchNames) {
		h.Branch = nil
	}

	return h
}

func santizeMetaData(h api.Heartbeat) api.Heartbeat {
	h.CursorPosition = nil
	h.Dependencies = nil
	h.LineNumber = nil
	h.Lines = nil

	return h
}

func shouldObfuscate(test string, patterns []*regexp.Regexp) bool {
	for _, p := range patterns {
		if p.Match([]byte(test)) {
			return true
		}
	}

	return false
}
