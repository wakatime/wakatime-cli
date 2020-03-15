package utils

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

// ShouldObfuscateProject Returns True if hide-project-names is true or the entity file path matches one in the list of obfuscated project paths.
func ShouldObfuscateProject(entity string, hideProjectNames []string) bool {
	for _, pattern := range hideProjectNames {
		b, err := strconv.ParseBool(pattern)
		if err == nil {
			return b
		}

		re, err := regexp.Compile("(?i)" + pattern)
		if err != nil {
			log.Printf("Regex error (%s) for hide-project-names pattern: %s", err, pattern)
			continue
		}

		if re.MatchString(entity) {
			return true
		}
	}

	return false
}

// ShouldObfuscateFilename Returns True if hide-file-names is true or the entity file path matches one in the list of obfuscated file paths.
func ShouldObfuscateFilename(entity string, hideFileNames []string) bool {
	for _, pattern := range hideFileNames {
		b, err := strconv.ParseBool(pattern)
		if err == nil {
			return b
		}

		re, err := regexp.Compile("(?i)" + pattern)
		if err != nil {
			log.Printf("Regex error (%s) for hide-file-names pattern: %s", err, pattern)
			continue
		}

		if re.MatchString(entity) {
			return true
		}
	}

	return false
}

// ShouldObfuscateBranch Returns True if hide-branch-names is true or the entity file path matches one in the list of obfuscated file paths.
func ShouldObfuscateBranch(entity string, branch string, hideBranchNames []string, default_ bool) bool {
	// when project names or file names are hidden and hide_branch_names is
	// not set, we default to hiding branch names along with file/project.
	if !default_ && len(hideBranchNames) == 0 {
		return true
	}

	if len(strings.TrimSpace(branch)) == 0 && len(hideBranchNames) == 0 {
		return false
	}

	for _, pattern := range hideBranchNames {
		b, err := strconv.ParseBool(pattern)
		if err == nil {
			return b
		}

		re, err := regexp.Compile("(?i)" + pattern)
		if err != nil {
			log.Printf("Regex error (%s) for hide-branch-names pattern: %s", err, pattern)
			continue
		}

		if re.MatchString(entity) || re.MatchString(branch) {
			return true
		}
	}

	return false
}
