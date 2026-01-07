//go:build ignore

// This program generates language_generated.go by reading all lexers from Chroma's
// GlobalLexerRegistry and creating stringToLanguage and chromaToLanguage maps.
// It also reads language.go to extract existing Language constants and string constants.
// Additionally, it updates language_test.go with languageTests() and languageTestsAliases().
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
	"unicode"

	"github.com/alecthomas/chroma/v2/lexers"
)

// languageInfo holds information about a language constant and its string value.
type languageInfo struct {
	constName string // e.g., "LanguageGo"
	strConst  string // e.g., "languageGoStr"
	strValue  string // e.g., "Go"
}

// chromaPriorityInfo holds a chroma lexer name to Language constant override.
type chromaPriorityInfo struct {
	chromaName string // e.g., "GAS"
	langConst  string // e.g., "LanguageAssembly"
}

func main() {
	// Read existing language constants and string constants from language.go
	existingLanguages, _, _, priorities, chromaPriorities := readLanguageFile()

	// Get all lexers from Chroma
	chromaLexers := lexers.GlobalLexerRegistry.Lexers

	// Maps to store forward and reverse mappings
	stringToLanguage := make(map[string]string)     // normalized string -> Language constant
	chromaToLanguage := make(map[string]string)     // normalized chroma name -> Language constant
	languageToString := make(map[string]string)     // Language constant -> string constant
	languageToChroma := make(map[string]string)     // Language constant -> raw chroma name (for StringChroma())
	newLanguages := make(map[string]*languageInfo)  // New languages to be added

	// First, populate from existing languages (only those with valid string constants)
	for langConst, info := range existingLanguages {
		if info.strConst == "" || info.strValue == "" {
			// Skip languages without string constants
			continue
		}
		// Skip LanguageUnknown - it should not be in stringToLanguage or languageToString
		if langConst == "LanguageUnknown" {
			continue
		}
		normalized := normalizeString(info.strValue)
		if normalized != "" {
			stringToLanguage[normalized] = langConst
		}
		languageToString[langConst] = info.strConst
	}

	// Track which languages we found matches for
	lexerToLanguage := make(map[string]string) // chroma lexer name -> Language constant
	primaryLexerNames := make(map[string]bool) // normalized primary lexer names (not aliases)

	// First pass: process all Chroma lexer names (not aliases) to establish primary mappings
	for _, lexer := range chromaLexers {
		config := lexer.Config()
		chromaName := config.Name

		// Try to find matching Language constant
		langConst := findMatchingLanguage(chromaName, existingLanguages)

		if langConst == "" {
			// Check aliases
			for _, alias := range config.Aliases {
				langConst = findMatchingLanguage(alias, existingLanguages)
				if langConst != "" {
					break
				}
			}
		}

		if langConst == "" {
			// New language from Chroma - generate constant name
			langConst = generateLanguageConstant(chromaName)
			strConst := generateStringConstant(chromaName)
			newLanguages[langConst] = &languageInfo{
				constName: langConst,
				strConst:  strConst,
				strValue:  chromaName,
			}
		}

		lexerToLanguage[chromaName] = langConst

		// Track this as a primary lexer name
		normalized := normalizeString(chromaName)
		primaryLexerNames[normalized] = true

		// Add lexer name to stringToLanguage (only if we have an existing language constant)
		if normalized != "" {
			if _, isNew := newLanguages[langConst]; !isNew {
				// Primary lexer name always takes precedence
				stringToLanguage[normalized] = langConst
			}
		}

		// Add to chromaToLanguage (only if different from standard name and is an existing language)
		chromaNormalized := normalizeString(chromaName)
		if chromaNormalized != "" {
			if existingInfo, ok := existingLanguages[langConst]; ok {
				// Check if chroma name differs from standard name
				standardNormalized := normalizeString(existingInfo.strValue)
				if chromaNormalized != standardNormalized {
					// Primary lexer name always takes precedence
					chromaToLanguage[chromaNormalized] = langConst
				}
				// Always populate languageToChroma with the raw chroma name for StringChroma()
				if _, exists := languageToChroma[langConst]; !exists {
					languageToChroma[langConst] = chromaName
				}
			}
		}
	}

	// Second pass: process aliases (only add if not already mapped by a primary lexer name)
	for _, lexer := range chromaLexers {
		config := lexer.Config()
		chromaName := config.Name
		langConst := lexerToLanguage[chromaName]

		// Add aliases to stringToLanguage (only for existing languages, only if not already mapped)
		for _, alias := range config.Aliases {
			normalizedAlias := normalizeString(alias)
			if normalizedAlias != "" {
				if _, isNew := newLanguages[langConst]; !isNew {
					if _, exists := stringToLanguage[normalizedAlias]; !exists {
						stringToLanguage[normalizedAlias] = langConst
					}
				}
			}
		}

		// Add aliases to chromaToLanguage (only for existing languages, only if not already mapped, and not a primary lexer name)
		for _, alias := range config.Aliases {
			aliasNormalized := normalizeString(alias)
			if aliasNormalized != "" {
				// Skip if this alias is a primary lexer name for another language
				if primaryLexerNames[aliasNormalized] {
					continue
				}
				if _, exists := chromaToLanguage[aliasNormalized]; !exists {
					// Only add if different from the standard name and is an existing language
					if existingInfo, ok := existingLanguages[langConst]; ok {
						standardNormalized := normalizeString(existingInfo.strValue)
						if aliasNormalized != standardNormalized {
							chromaToLanguage[aliasNormalized] = langConst
						}
					}
				}
			}
		}
	}

	// Apply priority overrides from language.go
	// The first priority entry for a Language sets the languageToString, subsequent ones only affect stringToLanguage
	seenLanguages := make(map[string]bool)
	for _, priority := range priorities {
		stringToLanguage[normalizeString(priority.strValue)] = priority.constName
		if !seenLanguages[priority.constName] {
			languageToString[priority.constName] = priority.strConst
			seenLanguages[priority.constName] = true
		}
	}

	// Apply chromaPriority overrides - these specify which Language to use for specific Chroma lexer names
	for _, cp := range chromaPriorities {
		normalized := normalizeString(cp.chromaName)
		if normalized != "" {
			chromaToLanguage[normalized] = cp.langConst
			// Also add to stringToLanguage ONLY if the chroma name doesn't already map to a different Language
			// For example: "GAS" should not be added because it already maps to LanguageGas
			// But "ApacheConf" should be added because it maps to LanguageApacheConfig (not LanguageApacheConf)
			if existingLang, exists := stringToLanguage[normalized]; !exists || existingLang == cp.langConst {
				stringToLanguage[normalized] = cp.langConst
			}
			// Update languageToChroma so StringChroma() returns the correct chroma name
			// But only if this Language doesn't already have a direct Chroma mapping
			if _, exists := languageToChroma[cp.langConst]; !exists {
				languageToChroma[cp.langConst] = cp.chromaName
			}
		}
	}

	// Populate languageToString for existing languages (not already set by priorities)
	for langConst, info := range existingLanguages {
		if info.strConst == "" {
			continue
		}
		// Skip LanguageUnknown - it should not be in languageToString
		if langConst == "LanguageUnknown" {
			continue
		}
		if _, exists := languageToString[langConst]; !exists {
			languageToString[langConst] = info.strConst
		}
	}

	// Generate output
	var buf bytes.Buffer
	buf.WriteString("// Code generated by go generate; DO NOT EDIT.\n")
	buf.WriteString("// Run 'go generate' to regenerate this file.\n\n")
	buf.WriteString("package heartbeat\n\n")

	// Generate stringToLanguage map
	buf.WriteString("// stringToLanguage maps normalized language strings to Language constants.\n")
	buf.WriteString("var stringToLanguage = map[string]Language{\n")
	var stringKeys []string
	for k := range stringToLanguage {
		stringKeys = append(stringKeys, k)
	}
	sort.Strings(stringKeys)
	for _, k := range stringKeys {
		fmt.Fprintf(&buf, "\t%q: %s,\n", k, stringToLanguage[k])
	}
	buf.WriteString("}\n\n")

	// Generate chromaToLanguage map
	buf.WriteString("// chromaToLanguage maps normalized chroma lexer names to Language constants.\n")
	buf.WriteString("// This map only contains entries where the chroma name differs from the standard name.\n")
	buf.WriteString("var chromaToLanguage = map[string]Language{\n")
	var chromaKeys []string
	for k := range chromaToLanguage {
		chromaKeys = append(chromaKeys, k)
	}
	sort.Strings(chromaKeys)
	for _, k := range chromaKeys {
		fmt.Fprintf(&buf, "\t%q: %s,\n", k, chromaToLanguage[k])
	}
	buf.WriteString("}\n\n")

	// Generate languageToString map
	buf.WriteString("// languageToString maps Language constants to their canonical string representation.\n")
	buf.WriteString("var languageToString = map[Language]string{\n")
	var langToStrKeys []string
	for k := range languageToString {
		langToStrKeys = append(langToStrKeys, k)
	}
	sort.Strings(langToStrKeys)
	for _, k := range langToStrKeys {
		fmt.Fprintf(&buf, "\t%s: %s,\n", k, languageToString[k])
	}
	buf.WriteString("}\n\n")

	// Generate languageToChroma map
	buf.WriteString("// languageToChroma maps Language constants to their chroma lexer name.\n")
	buf.WriteString("// Used by StringChroma() to return the exact chroma lexer name.\n")
	buf.WriteString("var languageToChroma = map[Language]string{\n")
	var langToChromaKeys []string
	for k := range languageToChroma {
		langToChromaKeys = append(langToChromaKeys, k)
	}
	sort.Strings(langToChromaKeys)
	for _, k := range langToChromaKeys {
		fmt.Fprintf(&buf, "\t%s: %q,\n", k, languageToChroma[k])
	}
	buf.WriteString("}\n")

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting generated code: %v\n", err)
		fmt.Fprintf(os.Stderr, "Unformatted code:\n%s\n", buf.String())
		os.Exit(1)
	}

	// Write to file
	err = os.WriteFile("language_generated.go", formatted, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing language_generated.go: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated language_generated.go with:\n")
	fmt.Printf("  - %d stringToLanguage entries\n", len(stringToLanguage))
	fmt.Printf("  - %d chromaToLanguage entries\n", len(chromaToLanguage))
	fmt.Printf("  - %d languageToString entries\n", len(languageToString))
	fmt.Printf("  - %d languageToChroma entries\n", len(languageToChroma))

	// Generate test entries
	// languageTests: map from language string value -> Language constant (tests Language.String() round-trip)
	// languageTestsAliases: map from alias string -> Language constant (tests that aliases parse correctly)
	languageTestsMap := make(map[string]string)        // string value -> Language constant
	languageTestsAliasesMap := make(map[string]string) // alias string -> Language constant

	// Build reverse map: Language constant -> primary string value
	// First from existingLanguages, then override with priority entries
	langToPrimaryStr := make(map[string]string)
	for langConst, info := range existingLanguages {
		if info.strValue != "" {
			langToPrimaryStr[langConst] = info.strValue
		}
	}
	// Apply priority overrides - the first priority entry for each Language sets the primary string
	priorityApplied := make(map[string]bool)
	for _, priority := range priorities {
		if !priorityApplied[priority.constName] && priority.strValue != "" {
			langToPrimaryStr[priority.constName] = priority.strValue
			priorityApplied[priority.constName] = true
		}
	}

	// Populate languageTests from langToPrimaryStr (respects priority overrides)
	// This tests that the primary string value parses back to the correct Language
	for langConst, primaryStr := range langToPrimaryStr {
		if primaryStr != "" {
			// Skip LanguageUnknown - it should not be tested
			if langConst == "LanguageUnknown" {
				continue
			}
			languageTestsMap[primaryStr] = langConst
		}
	}

	// Build a set of normalized primary strings to identify aliases
	normalizedPrimaryStrings := make(map[string]bool)
	for _, primaryStr := range langToPrimaryStr {
		if primaryStr != "" {
			normalizedPrimaryStrings[normalizeString(primaryStr)] = true
		}
	}

	// Add non-primary string values from existingLanguages as aliases
	// (these are strings that were overridden by priority entries)
	for langConst, info := range existingLanguages {
		if info.strValue == "" {
			continue
		}
		primaryStr := langToPrimaryStr[langConst]
		if normalizeString(info.strValue) != normalizeString(primaryStr) {
			// This is not the primary string, add as alias
			languageTestsAliasesMap[info.strValue] = langConst
		}
	}

	// Populate languageTestsAliases from stringToLanguage
	// Include entries where the raw string differs from the primary string for that language
	// Use the actual stringToLanguage mapping (what ParseLanguage returns)
	for rawStr, langConst := range stringToLanguage {
		primaryStr := langToPrimaryStr[langConst]
		if primaryStr == "" {
			continue
		}
		// Skip if this is the normalized form of the primary string
		if normalizeString(primaryStr) == rawStr {
			continue
		}
		// This is an alias - find the original (non-normalized) string
		// We need to track original strings for the test
	}

	// Actually, we need to track original strings -> Language mappings for aliases
	// Let's collect all Chroma lexer names and aliases with their mappings
	for _, lexer := range chromaLexers {
		config := lexer.Config()
		chromaName := config.Name

		// Look up what this normalizes to in stringToLanguage
		normalized := normalizeString(chromaName)
		if langConst, exists := stringToLanguage[normalized]; exists {
			primaryStr := langToPrimaryStr[langConst]
			if primaryStr != "" && normalizeString(chromaName) != normalizeString(primaryStr) {
				languageTestsAliasesMap[chromaName] = langConst
			}
		}

		// Add aliases
		for _, alias := range config.Aliases {
			normalized := normalizeString(alias)
			if langConst, exists := stringToLanguage[normalized]; exists {
				primaryStr := langToPrimaryStr[langConst]
				if primaryStr != "" && normalizeString(alias) != normalizeString(primaryStr) {
					languageTestsAliasesMap[alias] = langConst
				}
			}
		}
	}

	// Add priority aliases (from priority section in language.go)
	for _, priority := range priorities {
		primaryStr := langToPrimaryStr[priority.constName]
		if priority.strValue != "" && primaryStr != "" && normalizeString(priority.strValue) != normalizeString(primaryStr) {
			languageTestsAliasesMap[priority.strValue] = priority.constName
		}
	}

	// Add chromaPriority aliases - use the langConst from chromaPriority
	// But skip if this chromaName is already a primary string in languageTests
	for _, cp := range chromaPriorities {
		primaryStr := langToPrimaryStr[cp.langConst]
		if primaryStr != "" && normalizeString(cp.chromaName) != normalizeString(primaryStr) {
			// Skip if chromaName is already in languageTests (avoid duplicate test names)
			if _, exists := languageTestsMap[cp.chromaName]; !exists {
				languageTestsAliasesMap[cp.chromaName] = cp.langConst
			}
		}
	}

	// Also remove any aliases that conflict with primary strings
	for alias := range languageTestsAliasesMap {
		if _, exists := languageTestsMap[alias]; exists {
			delete(languageTestsAliasesMap, alias)
		}
	}

	// Update language_test.go
	updateTestFile(languageTestsMap, languageTestsAliasesMap)

	// Print new languages that need to be added to language.go
	if len(newLanguages) > 0 {
		fmt.Println("\n=== NEW LANGUAGES FROM CHROMA (need to be added to language.go) ===")
		var newLangKeys []string
		for k := range newLanguages {
			newLangKeys = append(newLangKeys, k)
		}
		sort.Strings(newLangKeys)
		for _, k := range newLangKeys {
			info := newLanguages[k]
			fmt.Printf("// %s represents the %s programming language.\n", info.constName, info.strValue)
			fmt.Printf("%s\n", info.constName)
		}
		fmt.Println("\n=== STRING CONSTANTS ===")
		for _, k := range newLangKeys {
			info := newLanguages[k]
			fmt.Printf("%s = %q\n", info.strConst, info.strValue)
		}
	}
}

// readLanguageFile reads language.go and extracts existing Language constants,
// string constants, chroma string constants, priority overrides, and chroma priority overrides.
func readLanguageFile() (map[string]*languageInfo, map[string]string, map[string]string, []*languageInfo, []*chromaPriorityInfo) {
	file, err := os.Open("language.go")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	existingLanguages := make(map[string]*languageInfo) // Language constant -> info
	stringConstants := make(map[string]string)          // string constant name -> value
	chromaStrConstants := make(map[string]string)       // chroma string constant name -> value
	var priorities []*languageInfo
	var chromaPriorities []*chromaPriorityInfo

	// Patterns
	languageConstPattern := regexp.MustCompile(`^\s*//\s*Language(\w+)\s+represent`)
	stringConstPattern := regexp.MustCompile(`^\s*(language\w+)\s*=\s*"([^"]+)"`)
	chromaStrConstPattern := regexp.MustCompile(`^\s*(language\w+ChromaStr)\s*=\s*"([^"]+)"`)
	priorityPattern := regexp.MustCompile(`^//\s*\d+\.\s*(Language\w+)\s*->\s*(language\w+)`)
	chromaPriorityPattern := regexp.MustCompile(`^//\s*\d+\.\s*(.+?)\s*->\s*(Language\w+)`)

	// State tracking
	inPriority := false
	inChromaPriority := false
	var currentLanguageConst string

	for scanner.Scan() {
		line := scanner.Text()

		// Check for priority section
		if strings.Contains(line, "// priority:") && !strings.Contains(line, "chromaPriority") {
			inPriority = true
			inChromaPriority = false
			continue
		}

		// Check for chromaPriority section
		if strings.Contains(line, "// chromaPriority:") {
			inChromaPriority = true
			inPriority = false
			continue
		}

		// Parse priority comments
		if inPriority {
			if matches := priorityPattern.FindStringSubmatch(line); matches != nil {
				langConst := matches[1]
				strConst := matches[2]
				// We'll resolve the string value later
				priorities = append(priorities, &languageInfo{
					constName: langConst,
					strConst:  strConst,
				})
				continue
			}
			// End of priority section (empty line or non-comment line)
			if !strings.HasPrefix(strings.TrimSpace(line), "//") {
				inPriority = false
			}
		}

		// Parse chromaPriority comments
		if inChromaPriority {
			if matches := chromaPriorityPattern.FindStringSubmatch(line); matches != nil {
				chromaName := strings.TrimSpace(matches[1])
				langConst := matches[2]
				chromaPriorities = append(chromaPriorities, &chromaPriorityInfo{
					chromaName: chromaName,
					langConst:  langConst,
				})
				continue
			}
			// End of chromaPriority section (empty line or non-comment line)
			if !strings.HasPrefix(strings.TrimSpace(line), "//") {
				inChromaPriority = false
			}
		}

		// Parse Language constant comments
		if matches := languageConstPattern.FindStringSubmatch(line); matches != nil {
			currentLanguageConst = "Language" + matches[1]
			continue
		}

		// Parse Language constant declarations (the line after the comment)
		if currentLanguageConst != "" {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, currentLanguageConst) {
				// This is the constant declaration line
				existingLanguages[currentLanguageConst] = &languageInfo{
					constName: currentLanguageConst,
				}
			}
			if !strings.HasPrefix(trimmed, "//") {
				currentLanguageConst = ""
			}
		}

		// Parse string constants
		if matches := stringConstPattern.FindStringSubmatch(line); matches != nil {
			constName := matches[1]
			constValue := matches[2]
			stringConstants[constName] = constValue
		}

		// Parse chroma string constants
		if matches := chromaStrConstPattern.FindStringSubmatch(line); matches != nil {
			constName := matches[1]
			constValue := matches[2]
			chromaStrConstants[constName] = constValue
		}
	}

	// Now link string constants to language constants
	for langConst, info := range existingLanguages {
		// Generate expected string constant name
		strConstName := generateStringConstantFromLanguage(langConst)
		if strValue, ok := stringConstants[strConstName]; ok {
			info.strConst = strConstName
			info.strValue = strValue
		}
	}

	// Resolve priority string values
	for _, p := range priorities {
		if strValue, ok := stringConstants[p.strConst]; ok {
			p.strValue = strValue
		}
	}

	return existingLanguages, stringConstants, chromaStrConstants, priorities, chromaPriorities
}

// findMatchingLanguage finds a Language constant that matches the given name.
func findMatchingLanguage(name string, existing map[string]*languageInfo) string {
	normalized := normalizeString(name)
	if normalized == "" {
		return ""
	}

	for langConst, info := range existing {
		// Skip languages without valid string values
		if info.strConst == "" || info.strValue == "" {
			continue
		}
		if normalizeString(info.strValue) == normalized {
			return langConst
		}
		// Also check if the constant name matches
		constNormalized := normalizeString(strings.TrimPrefix(langConst, "Language"))
		if constNormalized == normalized {
			return langConst
		}
	}

	return ""
}

// findChromaStringConstant finds a chroma string constant for the given name.
func findChromaStringConstant(name string, chromaStrConstants map[string]string) (string, bool) {
	normalized := normalizeString(name)

	for constName, constValue := range chromaStrConstants {
		if normalizeString(constValue) == normalized {
			return constName, true
		}
	}

	return "", false
}

// generateLanguageConstant generates a Language constant name from a Chroma name.
func generateLanguageConstant(name string) string {
	// Remove special characters and convert to PascalCase
	name = strings.ReplaceAll(name, "+", "Plus")
	name = strings.ReplaceAll(name, "#", "Sharp")
	name = strings.ReplaceAll(name, "*", "Star")

	var result strings.Builder
	result.WriteString("Language")

	words := splitIntoWords(name)
	for _, word := range words {
		if len(word) > 0 {
			result.WriteString(strings.ToUpper(string(word[0])))
			if len(word) > 1 {
				result.WriteString(word[1:])
			}
		}
	}

	return result.String()
}

// generateStringConstant generates a string constant name from a Chroma name.
func generateStringConstant(name string) string {
	// Generate languageXxxStr format
	name = strings.ReplaceAll(name, "+", "Plus")
	name = strings.ReplaceAll(name, "#", "Sharp")
	name = strings.ReplaceAll(name, "*", "Star")

	var result strings.Builder
	result.WriteString("language")

	words := splitIntoWords(name)
	for _, word := range words {
		if len(word) > 0 {
			result.WriteString(strings.ToUpper(string(word[0])))
			if len(word) > 1 {
				result.WriteString(word[1:])
			}
		}
	}
	result.WriteString("Str")

	return result.String()
}

// generateStringConstantFromLanguage generates a string constant name from a Language constant.
func generateStringConstantFromLanguage(langConst string) string {
	// LanguageFoo -> languageFooStr
	if !strings.HasPrefix(langConst, "Language") {
		return ""
	}
	suffix := strings.TrimPrefix(langConst, "Language")
	if len(suffix) == 0 {
		return ""
	}
	return "language" + suffix + "Str"
}

// splitIntoWords splits a string into words.
func splitIntoWords(s string) []string {
	var words []string
	var current strings.Builder

	for i, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			// Check for camelCase boundary
			if i > 0 && unicode.IsUpper(r) && current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
			current.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' || r == '/' || r == '.' {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		}
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}

// normalizeString normalizes a string for comparison (same as in language.go).
func normalizeString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "#", "sharp")
	s = strings.ReplaceAll(s, "++", "pp")

	return s
}

// updateTestFile updates language_test.go with languageTests() and languageTestsAliases().
func updateTestFile(languageTests, languageTestsAliases map[string]string) {
	// Read the existing test file
	content, err := os.ReadFile("language_test.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading language_test.go: %v\n", err)
		return
	}

	// Generate new languageTests() function content
	var testsContent strings.Builder
	testsContent.WriteString("func languageTests() map[string]heartbeat.Language {\n")
	testsContent.WriteString("\treturn map[string]heartbeat.Language{\n")

	var testKeys []string
	for k := range languageTests {
		testKeys = append(testKeys, k)
	}
	// Sort case-insensitively for proper alphabetical order
	sort.Slice(testKeys, func(i, j int) bool {
		return strings.ToLower(testKeys[i]) < strings.ToLower(testKeys[j])
	})

	for _, k := range testKeys {
		fmt.Fprintf(&testsContent, "\t\t%q: heartbeat.%s,\n", k, languageTests[k])
	}
	testsContent.WriteString("\t}\n")
	testsContent.WriteString("}")

	// Generate new languageTestsAliases() function content
	var aliasesContent strings.Builder
	aliasesContent.WriteString("func languageTestsAliases() map[string]heartbeat.Language {\n")
	aliasesContent.WriteString("\treturn map[string]heartbeat.Language{\n")

	var aliasKeys []string
	for k := range languageTestsAliases {
		aliasKeys = append(aliasKeys, k)
	}
	// Sort case-insensitively for proper alphabetical order
	sort.Slice(aliasKeys, func(i, j int) bool {
		return strings.ToLower(aliasKeys[i]) < strings.ToLower(aliasKeys[j])
	})

	for _, k := range aliasKeys {
		fmt.Fprintf(&aliasesContent, "\t\t%q: heartbeat.%s,\n", k, languageTestsAliases[k])
	}
	aliasesContent.WriteString("\t}\n")
	aliasesContent.WriteString("}")

	// Replace the functions in the file content
	contentStr := string(content)

	// Replace languageTests() function
	testsPattern := regexp.MustCompile(`(?s)func languageTests\(\) map\[string\]heartbeat\.Language \{.*?\n\t\}\n\}`)
	contentStr = testsPattern.ReplaceAllString(contentStr, testsContent.String())

	// Replace languageTestsAliases() function
	aliasesPattern := regexp.MustCompile(`(?s)func languageTestsAliases\(\) map\[string\]heartbeat\.Language \{.*?\n\t\}\n\}`)
	contentStr = aliasesPattern.ReplaceAllString(contentStr, aliasesContent.String())

	// Format the code
	formatted, err := format.Source([]byte(contentStr))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting language_test.go: %v\n", err)
		fmt.Fprintf(os.Stderr, "Writing unformatted content for debugging\n")
		os.WriteFile("language_test.go", []byte(contentStr), 0644)
		return
	}

	// Write the updated file
	err = os.WriteFile("language_test.go", formatted, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing language_test.go: %v\n", err)
		return
	}

	fmt.Printf("\nUpdated language_test.go with:\n")
	fmt.Printf("  - %d languageTests entries\n", len(languageTests))
	fmt.Printf("  - %d languageTestsAliases entries\n", len(languageTestsAliases))
}
