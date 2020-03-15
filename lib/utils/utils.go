package utils

import (
	"fmt"
	"log"
	"regexp"
	"runtime"
	"strings"

	"github.com/wakatime/wakatime-cli/constants"
	"github.com/wakatime/wakatime-cli/lib/system"
)

// GetUserAgent Get user agent
func GetUserAgent(plugin string) string {
	ver := runtime.Version()

	userAgent := fmt.Sprintf("wakatime/%s (%s) %s", constants.Version, system.GetOSInfo(), ver)

	if len(strings.TrimSpace(plugin)) > 0 {
		userAgent = fmt.Sprintf("%s %s", userAgent, plugin)
	} else {
		userAgent = fmt.Sprintf("%s %s", userAgent, "Unknown/0")
	}

	return userAgent
}

// ShouldExcludeByPattern ShouldExcludeByPattern
func ShouldExcludeByPattern(entity string, include []string, exclude []string) *string {
	//Find a way when compiling to ignore case
	if len(strings.TrimSpace(entity)) > 0 {
		for _, pattern := range include {
			re, err := regexp.CompilePOSIX(pattern)
			if err != nil {
				log.Printf("Regex error (%s) for include pattern: %s", err, pattern)
				continue
			}

			if re.MatchString(entity) {
				return nil
			}
		}
		for _, pattern := range exclude {
			re, err := regexp.CompilePOSIX(pattern)
			if err != nil {
				log.Printf("Regex error (%s) for exclude pattern: %s", err, pattern)
				continue
			}

			if re.MatchString(entity) {
				return &pattern
			}
		}
	}

	return nil
}

// IsExcludedByMissingProjectFile IsExcludedByMissingProjectFile
func IsExcludedByMissingProjectFile(entity string, includeOnlyWithProjectFile bool) bool {
	if !includeOnlyWithProjectFile {
		return false
	}

	return FindProjectFile(entity) == nil
}
