package deps

import (
	"fmt"
	"io"
	"os"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/regex"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	jww "github.com/spf13/jwalterweatherman"
)

// maxDependencyLength defines the maximum allowed length of a dependency.
// Any dependency exceeding this length will be discarded.
const maxDependencyLength = 200

// Config contains configurations for dependency scanning.
type Config struct {
	// FilePatterns will be matched against a file entities name and if matching, will skip
	// dependency scanning.
	FilePatterns []regex.Regex
}

// DependencyParser is a dependency parser for a programming language.
type DependencyParser interface {
	Parse(reader io.ReadCloser, lexer chroma.Lexer) ([]string, error)
}

// WithDetection initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to detect dependencies
// inside the entity file of heartbeats of type FileType. Will prioritize
// local file if available.
func WithDetection(c Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			for n, h := range hh {
				if h.EntityType != heartbeat.FileType {
					continue
				}

				if h.Language == nil {
					continue
				}

				if heartbeat.ShouldSanitize(h.Entity, c.FilePatterns) {
					continue
				}

				filepath := h.Entity

				if h.LocalFile != "" {
					filepath = h.LocalFile
				}

				dependencies, err := Detect(filepath, *h.Language)
				if err != nil {
					jww.WARN.Printf("error detecting dependencies of heartbeat: %s", err)
					continue
				}

				hh[n].Dependencies = dependencies
			}

			return next(hh)
		}
	}
}

// Detect parses the dependencies from a heartbeat file of a specific language.
func Detect(filepath string, language heartbeat.Language) ([]string, error) {
	lexer := lexers.Get(language.StringChroma())
	if lexer == nil {
		return nil, fmt.Errorf("unable to detect lexer for language %q", language)
	}

	var parser DependencyParser

	switch language {
	case heartbeat.LanguageC:
		parser = &ParserC{}
	case heartbeat.LanguageElm:
		parser = &ParserElm{}
	case heartbeat.LanguageGo:
		parser = &ParserGo{}
	case heartbeat.LanguageHaxe:
		parser = &ParserHaxe{}
	case heartbeat.LanguageJava:
		parser = &ParserJava{}
	case heartbeat.LanguagePHP:
		parser = &ParserPHP{}
	case heartbeat.LanguagePython:
		parser = &ParserPython{}
	case heartbeat.LanguageRust:
		parser = &ParserRust{}
	case heartbeat.LanguageSwift:
		parser = &ParserSwift{}
	default:
		jww.DEBUG.Printf("parsing dependencies not supported for language %q", language)
		return nil, nil
	}

	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %s", filepath, err)
	}

	deps, err := parser.Parse(f, lexer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dependencies: %s", err)
	}

	return filterDependencies(deps), nil
}

func filterDependencies(deps []string) []string {
	var (
		results []string
		unique  = make(map[string]struct{})
	)

	for _, d := range deps {
		// filter duplicate
		if _, ok := unique[d]; ok {
			continue
		}

		// filter dependencies exceeding max length
		if len(d) > maxDependencyLength {
			continue
		}

		unique[d] = struct{}{}

		results = append(results, d)
	}

	return results
}
