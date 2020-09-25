package language

import (
	"fmt"
	"io/ioutil"
	"os"
	fp "path/filepath"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	jww "github.com/spf13/jwalterweatherman"
)

// Config contains configurations for language detection.
type Config struct {
	Alternative string
	LocalFile   string
	Override    string
}

// WithDetection initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to detect and add programming
// language info to heartbeats of entity type 'file'.
func WithDetection(config Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			for n, h := range hh {
				if h.EntityType == heartbeat.FileType {
					language, err := Detect(h.Entity)
					if err != nil {
						jww.ERROR.Printf("failed to detect language on file entity %q: %s", h.Entity, err)
						continue
					}

					if language != "" {
						hh[n].Language = heartbeat.String(language)
					}
				}
			}

			return next(hh)
		}
	}
}

// Detect detects the language of a specific file.
func Detect(filepath string) (string, error) {
	if lang, ok := detectByFileExtensionSpecialCases(filepath); ok {
		return standardize(lang), nil
	}

	lang := detectByFileExtension(filepath)
	if lang != "" {
		return standardize(lang), nil
	}

	return "", fmt.Errorf("could not detect the language of file %q", filepath)
}

// detectByFileExtensionSpecialCases detects the language by file extension for
// some special cases.
func detectByFileExtensionSpecialCases(filepath string) (string, bool) {
	dir, file := fp.Split(filepath)
	ext := strings.ToLower(fp.Ext(file))

	// nolint
	if strings.HasPrefix(ext, ".h") || strings.HasPrefix(ext, ".c") {
		if correspondingFileExists(filepath, ".c") {
			return "C", true
		}

		if correspondingFileExists(filepath, ".m") {
			return "Objective-C", true
		}

		if correspondingFileExists(filepath, ".mm") {
			return "Objective-C++", true
		}

		if folderContainsCPPFiles(dir) {
			return "C++", true
		}

		if folderContainsCFiles(dir) {
			return "C", true
		}
	}

	if ext == ".m" && correspondingFileExists(filepath, ".h") {
		return "Objective-C", true
	}

	if ext == ".mm" && correspondingFileExists(filepath, ".h") {
		return "Objective-C++", true
	}

	return "", false
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
func correspondingFileExists(filepath string, extension string) bool {
	_, file := fp.Split(filepath)
	ext := strings.ToLower(fp.Ext(file))
	noExtension := filepath[:len(filepath)-len(ext)]

	for _, ext := range []string{extension, strings.ToUpper(extension)} {
		if _, err := os.Stat(noExtension + ext); err == nil {
			return true
		}
	}

	return false
}

// detectByFileExtension detects the language by file extension.
func detectByFileExtension(filepath string) string {
	filepath = strings.ToLower(filepath)

	switch {
	case strings.HasSuffix(filepath, ".asp"):
		return "ASP"
	case strings.HasSuffix(filepath, ".bas") ||
		strings.HasSuffix(filepath, ".BAS"):
		return "Basic"
	case strings.HasSuffix(filepath, ".cfm"):
		return "ColdFusion"
	case strings.HasSuffix(filepath, ".cshtml"):
		return "CSHTML"
	case strings.HasSuffix(filepath, ".pas"):
		return "Delphi"
	case strings.HasSuffix(filepath, ".fs"):
		return "F#"
	case strings.HasSuffix(filepath, "go.mod"):
		return "Go"
	case strings.HasSuffix(filepath, ".gs") ||
		strings.HasSuffix(filepath, ".gsp") ||
		strings.HasSuffix(filepath, ".gst") ||
		strings.HasSuffix(filepath, ".gsx"):
		return "Gosu"
	case strings.HasSuffix(filepath, ".haml"):
		return "Haml"
	case strings.HasSuffix(filepath, ".jade"):
		return "Jade"
	case strings.HasSuffix(filepath, ".mjs"):
		return "JavaScript"
	case strings.HasSuffix(filepath, ".jsx"):
		return "JSX"
	case strings.HasSuffix(filepath, ".less"):
		return "LESS"
	case strings.HasSuffix(filepath, ".mustache"):
		return "Mustache"
	case strings.HasSuffix(filepath, ".mm"):
		return "Objective-C"
	// Directly matching Typescript here, to emulate the behavior of python's
	// pygments library, which exclusively associate typescript with *.ts files.
	case strings.HasSuffix(filepath, ".ts"):
		return "TypeScript"
	// Directly matching Typoscript here, to emulate the behavior of python's
	// pygments library, which does associate typoscript with  *.typoscript files.
	case strings.HasSuffix(filepath, ".typoscript"):
		return "TypoScript"
	case strings.HasSuffix(filepath, ".xaml"):
		return "XAML"
	}

	lexer := Match(filepath)
	if lexer == nil {
		return ""
	}

	return lexer.Config().Name
}

// loadFolderExtensions loads all existing from a folder.
func loadFolderExtensions(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %s", err)
	}

	var extensions []string

	for _, f := range files {
		_, file := fp.Split(f.Name())
		extensions = append(extensions, strings.ToLower(fp.Ext(file)))
	}

	return extensions, nil
}

// standardize converts a language string to the standardized name.
func standardize(lang string) string {
	switch lang {
	case "EmacsLisp":
		return "Emacs Lisp"
	case "GAS":
		return "Assembly"
	case "markdown":
		return "Markdown"
	case "FSharp":
		return "F#"
	}

	return lang
}
