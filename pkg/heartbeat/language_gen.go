//go:build ignore

// This program generates language_generated.go by reading language_maps.go
// and creating reverse mappings (Language -> string) from the forward mappings.
// Run it with: go generate
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"os"
	"regexp"
	"sort"
	"strings"
)

func main() {
	// Read language_maps.go to extract the mappings
	file, err := os.Open("language_maps.go")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Maps to store mappings
	languageToStringMap := make(map[string]string)
	languageToChromaMap := make(map[string]string)

	// Maps to store priority overrides from comments
	stringPriorityMap := make(map[string]string)
	chromaPriorityMap := make(map[string]string)

	// Track Languages that should not have chroma entries (multiple aliases, none canonical)
	chromaExcluded := make(map[string]bool)

	// Pattern to match map entries: normalizeString(languageXxxStr): LanguageXxx,
	entryPattern := regexp.MustCompile(`^\s*normalizeString\((language\w+Str)\):\s*(Language\w+),`)

	// Pattern to match priority comments: // N. LanguageXxx -> languageXxxStr
	priorityPattern := regexp.MustCompile(`^//\s*\d+\.\s*(Language\w+)\s*->\s*(language\w+Str)`)

	inStringMap := false
	inChromaMap := false
	inStringPriority := false
	inChromaPriority := false

	for scanner.Scan() {
		line := scanner.Text()

		// Detect priority comment section for stringToLanguage
		if strings.Contains(line, "// priority:") {
			if !inStringMap && !inChromaMap {
				// We're before the stringToLanguage map
				inStringPriority = true
				inChromaPriority = false
			} else if inChromaMap {
				// This would be a priority section for chroma (if it exists)
				inChromaPriority = true
				inStringPriority = false
			}
			continue
		}

		// Parse priority comments
		if inStringPriority || inChromaPriority {
			if matches := priorityPattern.FindStringSubmatch(line); matches != nil {
				language := matches[1]
				strConst := matches[2]
				if inStringPriority {
					stringPriorityMap[language] = strConst
				} else {
					chromaPriorityMap[language] = strConst
				}
				continue
			}
			// End of priority section when we hit a non-priority line (like nolint or blank)
			if !strings.HasPrefix(strings.TrimSpace(line), "//") || strings.Contains(line, "nolint") {
				inStringPriority = false
				inChromaPriority = false
			}
		}

		// Detect map boundaries
		if regexp.MustCompile(`^var stringToLanguage\s*=`).MatchString(line) {
			inStringMap = true
			inChromaMap = false
			inStringPriority = false
			continue
		}
		if regexp.MustCompile(`^var chromaToLanguage\s*=`).MatchString(line) {
			inChromaMap = true
			inStringMap = false
			inChromaPriority = false
			continue
		}
		if (inStringMap || inChromaMap) && regexp.MustCompile(`^}`).MatchString(line) {
			inStringMap = false
			inChromaMap = false
			continue
		}

		if matches := entryPattern.FindStringSubmatch(line); matches != nil {
			strConst := matches[1]
			language := matches[2]

			if inStringMap {
				// Check if there's a priority override for this language
				if priorityStr, hasPriority := stringPriorityMap[language]; hasPriority {
					// Use priority value
					languageToStringMap[language] = priorityStr
				} else if _, exists := languageToStringMap[language]; !exists {
					// Only keep first occurrence (canonical string for this Language)
					languageToStringMap[language] = strConst
				}
			} else if inChromaMap {
				// Check if there's a priority override for this language
				if priorityStr, hasPriority := chromaPriorityMap[language]; hasPriority {
					// Use priority value
					languageToChromaMap[language] = priorityStr
				} else if chromaExcluded[language] {
					// Already determined this Language should not have a chroma entry
					// (multiple aliases, none canonical)
					continue
				} else if _, exists := languageToChromaMap[language]; !exists {
					// First occurrence - store it
					languageToChromaMap[language] = strConst
				} else {
					// Multiple chroma names map to same Language
					// Only keep entries where constant name matches Language type (canonical)
					// e.g., languageAMPLChromaStr -> LanguageAMPL (canonical, keep)
					// but languageGoHTMLTemplateChromaStr -> LanguageGo (alias, remove)
					existingConst := languageToChromaMap[language]
					existingExpected := "Language" + strings.TrimSuffix(strings.TrimPrefix(existingConst, "language"), "ChromaStr")
					newExpected := "Language" + strings.TrimSuffix(strings.TrimPrefix(strConst, "language"), "ChromaStr")

					// If neither matches, remove the entry and mark as excluded
					if existingExpected != language && newExpected != language {
						delete(languageToChromaMap, language)
						chromaExcluded[language] = true
					} else if newExpected == language && existingExpected != language {
						// New one matches, replace
						languageToChromaMap[language] = strConst
					}
					// If existing matches and new doesn't, keep existing (do nothing)
					// If both match, keep existing (do nothing)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading language_maps.go: %v\n", err)
		os.Exit(1)
	}

	// Sort keys for consistent output
	var stringKeys []string
	for k := range languageToStringMap {
		stringKeys = append(stringKeys, k)
	}
	sort.Strings(stringKeys)

	var chromaKeys []string
	for k := range languageToChromaMap {
		chromaKeys = append(chromaKeys, k)
	}
	sort.Strings(chromaKeys)

	// Generate output
	var buf bytes.Buffer
	buf.WriteString("// Code generated by go generate; DO NOT EDIT.\n")
	buf.WriteString("// Run 'go generate' to regenerate this file.\n\n")
	buf.WriteString("package heartbeat\n\n")
	buf.WriteString("// languageToString maps Language constants to their canonical string representation.\n")
	buf.WriteString("var languageToString = map[Language]string{\n")
	for _, lang := range stringKeys {
		fmt.Fprintf(&buf, "\t%s: %s,\n", lang, languageToStringMap[lang])
	}
	buf.WriteString("}\n\n")
	buf.WriteString("// languageToChroma maps Language constants to their chroma lexer name.\n")
	buf.WriteString("// Only includes languages where chroma name differs from standard name.\n")
	buf.WriteString("var languageToChroma = map[Language]string{\n")
	for _, lang := range chromaKeys {
		fmt.Fprintf(&buf, "\t%s: %s,\n", lang, languageToChromaMap[lang])
	}
	buf.WriteString("}\n")

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting generated code: %v\n", err)
		os.Exit(1)
	}

	// Write to file
	err = os.WriteFile("language_generated.go", formatted, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing language_generated.go: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated language_generated.go with %d string mappings and %d chroma mappings\n",
		len(stringKeys), len(chromaKeys))
}
