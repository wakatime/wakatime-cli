package ini

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Key holds a single INI section and key.
type Key struct {
	// Section is an INI section for this key
	Section string
	// Name is the key name of an INI setting
	Name string
}

// GetKey returns the value for given INI Key.
func GetKey(iniFile string, key Key) string {
	keys := []Key{key}

	return GetKeys(iniFile, keys)[key]
}

// GetKeys returns the values for given INI Keys.
func GetKeys(iniFile string, keys []Key) map[Key]string {
	result := map[Key]string{}
	finding := map[Key]bool{}
	found := map[Key]bool{}

	for _, key := range keys {
		finding[key] = true
	}

	fh, err := os.Open(iniFile)
	if err != nil {
		return result
	}

	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	scanner.Split(bufio.ScanLines)

	var (
		currentSection string
		currentKey     Key
		multiline      []string
	)

	for scanner.Scan() {
		line := scanner.Text()

		if len(multiline) > 0 {
			if isMultiline(line) {
				multiline = append(multiline, line)
				continue
			}

			if finding[currentKey] {
				result[currentKey] = strings.Join(multiline, "\n")
				found[currentKey] = true

				// return early if we've already found all they keys
				if len(found) == len(finding) {
					return result
				}
			}

			multiline = []string{}
		}

		if isSection(line) {
			currentSection = getSectionName(line)
			continue
		}

		split := strings.SplitN(line, "=", 2)
		if len(split) != 2 {
			continue
		}

		possibleKey := getPossibleKeyName(split)
		if possibleKey == "" {
			continue
		}

		currentKey = Key{
			Section: currentSection,
			Name:    possibleKey,
		}

		multiline = append(multiline, strings.TrimLeft(strings.TrimLeft(split[1], " "), "\t"))
	}

	if len(multiline) > 0 && finding[currentKey] {
		result[currentKey] = strings.Join(multiline, "\n")
		found[currentKey] = true
	}

	// key not found in INI file, return empty string
	return result
}

// SetKey saves the value for given INI Key.
func SetKey(iniFile string, key Key, value string) error {
	keys := map[Key]string{key: value}
	return SetKeys(iniFile, keys)
}

// SetKeys saves multiple values for given INI Keys.
func SetKeys(iniFile string, keys map[Key]string) error {
	if len(keys) == 0 {
		return nil
	}

	fh, err := os.Open(iniFile)
	if err != nil {
		return fmt.Errorf("failed to open ini file %q: %s", iniFile, err)
	}

	scanner := bufio.NewScanner(fh)
	scanner.Split(bufio.ScanLines)

	var (
		lines          []string
		currentSection string
	)

	found := map[Key]bool{}

	for scanner.Scan() {
		line := scanner.Text()

		if isSection(line) { // nolint:nestif
			for key, value := range keys {
				if found[key] {
					continue
				}

				if currentSection == key.Section {
					found[key] = true

					lines = append(lines, key.Name+" = "+value)

					break
				}
			}

			currentSection = getSectionName(line)
		} else {
			localFound := false

			for key, value := range keys {
				if found[key] {
					continue
				}

				if currentSection == key.Section {
					split := strings.SplitN(line, "=", 2)
					if len(split) == 2 {
						possibleKey := getPossibleKeyName(split)
						if possibleKey == key.Name {
							found[key] = true
							localFound = true
							lines = append(lines, key.Name+" = "+value)
							break
						}
					}
				}
			}

			if localFound {
				continue
			}
		}

		lines = append(lines, line)
	}

	for key, value := range keys {
		if found[key] {
			continue
		}

		if currentSection != key.Section {
			lines = append(lines, "["+key.Section+"]")

			currentSection = key.Section
		}

		lines = append(lines, key.Name+" = "+value)
	}

	fh.Close()

	output := strings.Join(lines, "\n")

	return ioutil.WriteFile(iniFile, []byte(output), 0644) // nolint:gosec
}

func isSection(line string) bool {
	return strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") && len(line) > 2
}

func isMultiline(line string) bool {
	return strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")
}

func getSectionName(line string) string {
	return strings.TrimSuffix(strings.TrimPrefix(line, "["), "]")
}

func getPossibleKeyName(splitLine []string) string {
	if len(splitLine) == 0 {
		return ""
	}

	return strings.TrimRight(strings.TrimRight(splitLine[0], " "), "\t")
}
