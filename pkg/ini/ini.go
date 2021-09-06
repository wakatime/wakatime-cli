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

// Item holds the value for a Key.
type Item struct {
	// Key is the section and key name for this INI setting
	Key Key
	// Value is the value for this Key, or empty string
	Value string
}

// GetKey returns the value for given INI Key.
func GetKey(iniFile string, key Key) (Item, error) {
	fh, err := os.Open(iniFile)
	if err != nil {
		return Item{}, fmt.Errorf("failed to open ini file %q: %s", iniFile, err)
	}

	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	scanner.Split(bufio.ScanLines)

	var (
		currentSection string
		multiline      []string
	)

	for scanner.Scan() {
		line := scanner.Text()

		if len(multiline) > 0 {
			if !isMultiline(line) {
				return Item{
					Key:   key,
					Value: strings.Join(multiline, "\n"),
				}, nil
			}

			multiline = append(multiline, line)

			continue
		}

		if isSection(line) {
			currentSection = getSectionName(line)
			continue
		}

		if currentSection != key.Section {
			continue
		}

		split := strings.SplitN(line, "=", 2)
		if len(split) != 2 {
			continue
		}

		possibleKey := getPossibleKeyName(split)
		if possibleKey != key.Name {
			continue
		}

		multiline = append(multiline, strings.TrimLeft(strings.TrimLeft(split[1], " "), "\t"))
	}

	// key not found in INI file, return empty string
	return Item{
		Key:   key,
		Value: "",
	}, nil
}

// SetKey saves the value for given INI Key.
func SetKey(iniFile string, value Item) error {
	return SetKeys(iniFile, []Item{value})
}

// SetKeys saves multiple values for given INI Keys.
func SetKeys(iniFile string, values []Item) error {
	if len(values) == 0 {
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
			for _, value := range values {
				if found[value.Key] {
					continue
				}

				if currentSection == value.Key.Section {
					found[value.Key] = true

					lines = append(lines, value.Key.Name+" = "+value.Value)

					break
				}
			}

			currentSection = getSectionName(line)
		} else {
			localFound := false

			for _, value := range values {
				if found[value.Key] {
					continue
				}

				if currentSection == value.Key.Section {
					split := strings.SplitN(line, "=", 2)
					if len(split) == 2 {
						possibleKey := getPossibleKeyName(split)
						if possibleKey == value.Key.Name {
							found[value.Key] = true
							localFound = true
							lines = append(lines, value.Key.Name+" = "+value.Value)
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

	for _, value := range values {
		if found[value.Key] {
			continue
		}

		if currentSection != value.Key.Section {
			lines = append(lines, "["+value.Key.Section+"]")

			currentSection = value.Key.Section
		}

		lines = append(lines, value.Key.Name+" = "+value.Value)
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
