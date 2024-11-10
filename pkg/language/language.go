package language

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// Config defines language detection options.
type Config struct {
	// GuessLanguage enables detecting lexer language from file contents.
	GuessLanguage bool
}

// WithDetection initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to detect and add programming
// language info to heartbeats of entity type 'file'.
func WithDetection(config Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(ctx context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			logger := log.Extract(ctx)
			logger.Debugln("execute language detection")

			for n, h := range hh {
				if hh[n].Language != nil {
					continue
				}

				filepath := h.Entity

				if h.LocalFile != "" {
					filepath = h.LocalFile
				}

				language, err := Detect(ctx, filepath, config.GuessLanguage)
				if err != nil && hh[n].LanguageAlternate != "" {
					hh[n].Language = heartbeat.PointerTo(hh[n].LanguageAlternate)

					continue
				}

				if err != nil {
					logger.Debugf("failed to detect language on file entity %q: %s", h.Entity, err)

					continue
				}

				hh[n].Language = heartbeat.PointerTo(language.String())
			}

			return next(ctx, hh)
		}
	}
}

// Detect detects the language of a specific file. If guessLanguage is true,
// Chroma will be used to detect a language from the file contents.
func Detect(ctx context.Context, fp string, guessLanguage bool) (heartbeat.Language, error) {
	if language, ok := detectSpecialCases(ctx, fp); ok {
		return language, nil
	}

	var language heartbeat.Language

	languageChroma, weight, ok := detectChromaCustomized(ctx, fp, guessLanguage)
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
func detectSpecialCases(ctx context.Context, fp string) (heartbeat.Language, bool) {
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

		if folderContainsCPPFiles(ctx, dir) {
			return heartbeat.LanguageCPP, true
		}

		if folderContainsCFiles(ctx, dir) {
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
func folderContainsCFiles(ctx context.Context, dir string) bool {
	logger := log.Extract(ctx)

	extensions, err := loadFolderExtensions(dir)
	if err != nil {
		logger.Warnf("failed loading folder extensions: %s", err)
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
func folderContainsCPPFiles(ctx context.Context, dir string) bool {
	logger := log.Extract(ctx)

	extensions, err := loadFolderExtensions(dir)
	if err != nil {
		logger.Warnf("failed loading folder extensions: %s", err)
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

// loadFolderExtensions loads all existing file extensions from a folder.
func loadFolderExtensions(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
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
