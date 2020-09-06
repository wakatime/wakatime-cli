package language

import (
	"fmt"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/lexers"
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
	lang := detectByFileExtension(filepath)
	if lang != "" {
		return standardize(lang), nil
	}

	return "", fmt.Errorf("could not detect the language of file %q", filepath)
}

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

	lexer := lexers.Match(filepath)
	if lexer == nil {
		return ""
	}

	return lexer.Config().Name
}

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
