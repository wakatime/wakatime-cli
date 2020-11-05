package language

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	jww "github.com/spf13/jwalterweatherman"
)

// Config contains configurations for language detection.
type Config struct {
	Alternate string
	Override  string
}

// WithDetection initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to detect and add programming
// language info to heartbeats of entity type 'file'.
func WithDetection(config Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			for n, h := range hh {
				if config.Override != "" {
					language, ok := heartbeat.ParseLanguage(config.Override)
					if !ok {
						jww.WARN.Printf("Failed to parse override language %q", config.Alternate)
					}

					hh[n].Language = language

					continue
				}

				filepath := h.Entity

				if h.LocalFile != "" {
					filepath = h.LocalFile
				}

				language, err := Detect(filepath)
				if err != nil {
					jww.ERROR.Printf("failed to detect language on file entity %q: %s", h.Entity, err)

					if config.Alternate != "" {
						parsed, ok := heartbeat.ParseLanguage(config.Alternate)
						if !ok {
							jww.WARN.Printf("Failed to parse alternate language %q", config.Alternate)
						}

						hh[n].Language = parsed
					}

					continue
				}

				hh[n].Language = language
			}

			return next(hh)
		}
	}
}

// Detect detects the language of a specific file.
func Detect(fp string) (heartbeat.Language, error) {
	if language, ok := detectSpecialCases(fp); ok {
		return language, nil
	}

	return heartbeat.LanguageUnknown, fmt.Errorf("could not detect the language of file %q", fp)
}

// detectSpecialCases detects the language by file extension for some special cases.
func detectSpecialCases(fp string) (heartbeat.Language, bool) {
	dir, file := filepath.Split(fp)
	ext := strings.ToLower(filepath.Ext(file))

	// nolint
	if strings.HasPrefix(ext, ".h") || strings.HasPrefix(ext, ".c") {
		if correspondingFileExists(fp, ".c") {
			return heartbeat.LanguageC, true
		}

		if correspondingFileExists(fp, ".m") {
			return heartbeat.LanguageObjectiveC, true
		}

		if correspondingFileExists(fp, ".mm") {
			return heartbeat.LanguageObjectiveCPP, true
		}

		if folderContainsCPPFiles(dir) {
			return heartbeat.LanguageCPP, true
		}

		if folderContainsCFiles(dir) {
			return heartbeat.LanguageC, true
		}
	}

	if ext == ".m" && correspondingFileExists(fp, ".h") {
		return heartbeat.LanguageObjectiveC, true
	}

	if ext == ".mm" && correspondingFileExists(fp, ".h") {
		return heartbeat.LanguageObjectiveCPP, true
	}

	return heartbeat.LanguageUnknown, false
}

// folderContainsCFiles returns true, if filder contains c files.
func folderContainsCFiles(dir string) bool {
	extensions, err := loadFolderExtensions(dir)
	if err != nil {
		jww.ERROR.Printf("failed loading folder extensions: %s", err)
		return false
	}

	for _, e := range extensions {
		if e == ".c" {
			return true
		}
	}

	return false
}

// folderContainsCFiles returns true, if filder contains c++ files.
func folderContainsCPPFiles(dir string) bool {
	extensions, err := loadFolderExtensions(dir)
	if err != nil {
		jww.ERROR.Printf("failed loading folder extensions: %s", err)
		return false
	}

	cppExtensions := []string{".cpp", ".hpp", ".c++", ".h++", ".cc", ".hh", ".cxx", ".hxx", ".C", ".H", ".cp", ".CPP"}
	for _, cppExt := range cppExtensions {
		for _, e := range extensions {
			if e == cppExt {
				return true
			}
		}
	}

	return false
}

// correspondingFileExists returns true if corresponding file with the provided extension exists.
// E.g. will return true, if called with "/tmp/file.go" and "txt" and /tmp/file.txt existis.
func correspondingFileExists(fp string, extension string) bool {
	_, file := filepath.Split(fp)
	ext := strings.ToLower(filepath.Ext(file))
	noExtension := fp[:len(fp)-len(ext)]

	for _, ext := range []string{extension, strings.ToUpper(extension)} {
		if _, err := os.Stat(noExtension + ext); err == nil {
			return true
		}
	}

	return false
}

// loadFolderExtensions loads all existing from a folder.
func loadFolderExtensions(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %s", err)
	}

	var extensions []string

	for _, f := range files {
		_, file := filepath.Split(f.Name())
		extensions = append(extensions, strings.ToLower(filepath.Ext(file)))
	}

	return extensions, nil
}
