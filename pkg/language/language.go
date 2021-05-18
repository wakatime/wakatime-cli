package language

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// WithDetection initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to detect and add programming
// language info to heartbeats of entity type 'file'.
func WithDetection() heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			log.Debugln("execute language detection")

			for n, h := range hh {
				if hh[n].Language != nil {
					continue
				}

				filepath := h.Entity

				if h.LocalFile != "" {
					filepath = h.LocalFile
				}

				language, err := Detect(filepath)
				if err != nil && hh[n].LanguageAlternate != "" {
					hh[n].Language = heartbeat.String(hh[n].LanguageAlternate)

					continue
				}

				if err != nil {
					log.Warnf("failed to detect language on file entity %q: %s", h.Entity, err)

					continue
				}

				hh[n].Language = heartbeat.String(language.String())
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

	var language heartbeat.Language

	languageChroma, weight, ok := detectChromaCustomized(fp)
	if ok {
		language = languageChroma
	}

	languageVim, weightVim, okVim := detectVimModeline(fp)
	if okVim && weightVim > weight {
		// use language from vim modeline, if weight is higher
		language = languageVim
	}

	if language == heartbeat.LanguageUnknown {
		return heartbeat.LanguageUnknown, fmt.Errorf("could not detect the language of file %q", fp)
	}

	return language, nil
}

// detectSpecialCases detects the language by file extension for some special cases.
func detectSpecialCases(fp string) (heartbeat.Language, bool) {
	dir, file := filepath.Split(fp)
	ext := strings.ToLower(filepath.Ext(file))

	switch file {
	case "go.mod":
		return heartbeat.LanguageGo, true
	case "CMmakeLists.txt":
		return heartbeat.LanguageCMake, true
	}

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
		log.Warnf("failed loading folder extensions: %s", err)
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
		log.Warnf("failed loading folder extensions: %s", err)
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
