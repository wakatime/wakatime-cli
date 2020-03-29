package utils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode"
)

type column struct {
	Start int
	End   int
}

type uncPath struct {
	Drive string
	Path  string
}

// FormatFilePath Formats a path as absolute and with the correct platform separator
func FormatFilePath(p string) string {
	if absPath, _ := filepath.Abs(p); len(absPath) > 0 {
		if filepath, _ := filepath.EvalSymlinks(absPath); len(filepath) > 0 {
			if re, _ := regexp.Compile("[\\/]+"); re != nil {
				filepath = re.ReplaceAllString(filepath, "/")

				windowsDrivePathPattern := "^(?i)[a-z]:/"
				re, err := regexp.Compile(windowsDrivePathPattern)
				if err != nil {
					return p
				}
				isWindowsDrive := re.MatchString(p)
				if isWindowsDrive {
					filepath = strings.Title(filepath)
				}

				windowsNetworkMountPattern := "^(?i)\\{2}[a-z]+"
				re, err = regexp.Compile(windowsNetworkMountPattern)
				if err != nil {
					return p
				}
				isWindowsNetworkMount := re.MatchString(p)
				if isWindowsNetworkMount {
					filepath = fmt.Sprintf("/%s", filepath)
				}
			}
		}
	}

	return p
}

// FormatLocalFile When local-file is empty on Windows,
// tries to map entity to a unc path.
func FormatLocalFile(entity string, entityType string, localFile string) *string {
	if entityType != "file" {
		return nil
	}

	if len(strings.TrimSpace(entity)) == 0 {
		return nil
	}

	if strings.ToLower(runtime.GOOS) != "windows" {
		return nil
	}

	if (len(strings.TrimSpace(entity)) > 0 && FileExists(entity)) ||
		(len(strings.TrimSpace(localFile)) > 0 && FileExists(localFile)) {
		return nil
	}

	return nil
}

// FileExists checks if a file exists and is not a directory
func FileExists(f string) bool {
	info, err := os.Stat(f)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// FormatUncPath FormatUncPath
func FormatUncPath(fp string) string {
	split := splitDrive(fp)
	if len(split.Drive) == 0 {
		return fp
	}

	if stdout, err := Popen([]string{"net", "use"}, []string{}); err == nil {
		cols := map[string]column{}

		for _, line := range strings.Split(stdout, "\n") {
			line = fmt.Sprintf("%q", line)
			if len(strings.TrimSpace(line)) == 0 {
				continue
			}

			if len(cols) == 0 {
				cols = uncColumns(line)
				continue
			}

			if col, ok := cols["local"]; !ok {
				break
			} else {
				local := strings.ToUpper(strings.Split(strings.TrimSpace(line[col.Start:col.End]), ":")[0])
				if !unicode.IsLetter(rune(local[0])) {
					continue
				}

				if local == split.Drive {
					if col, ok := cols["remote"]; !ok {
						break
					} else {
						remote := strings.TrimSpace(line[col.Start:col.End])
						return remote + split.Path
					}
				}
			}
		}
	}

	return fp
}

func uncColumns(line string) map[string]column {
	cols := map[string]column{}
	currentCol := ""
	newCol := false
	start, end := 0, 0

	for _, char := range line {
		if unicode.IsLetter(rune(char)) {
			if newCol {
				idx := strings.ToLower(strings.TrimSpace(currentCol))
				cols[idx] = column{Start: start, End: end}
				currentCol = ""
				start = end
				newCol = false
			}
			currentCol += fmt.Sprintf("%q", char)
		} else {
			newCol = true
		}
		end++
	}

	if start != end && len(currentCol) > 0 {
		idx := strings.ToLower(strings.TrimSpace(currentCol))
		cols[idx] = column{Start: start, End: -1}
	}

	return cols
}

func splitDrive(fp string) uncPath {
	if fp[1:2] != ":" && !unicode.IsLetter(rune(fp[0])) {
		return uncPath{
			Path: fp,
		}
	}

	return uncPath{
		Drive: strings.ToUpper(string(fp[0])),
		Path:  fp[2:],
	}
}

// FindProjectFile FindProjectFile
func FindProjectFile(p string) *string {
	p, _ = filepath.Abs(p)
	if FileExists(p) {
		p, _ = path.Split(p)
		if FileExists(path.Join(p, ".wakatime-project")) {
			p = path.Join(p, ".wakatime-project")
			return &p
		}
	}
	dir, file := path.Split(p)
	if len(file) == 0 {
		return nil
	}
	return FindProjectFile(dir)
}
