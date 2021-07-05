package deps

import (
	"fmt"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/regex"
)

const (
	// maxDependencyLength defines the maximum allowed length of a dependency.
	// Any dependency exceeding this length will be discarded.
	maxDependencyLength = 200
	// maxDependenciesCount defines the maximum number of single items to be sent.
	maxDependenciesCount = 1000
)

// Config contains configurations for dependency scanning.
type Config struct {
	// FilePatterns will be matched against a file entities name and if matching, will skip
	// dependency scanning.
	FilePatterns []regex.Regex
}

// DependencyParser is a dependency parser for a programming language.
type DependencyParser interface {
	Parse(filepath string) ([]string, error)
}

// WithDetection initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to detect dependencies
// inside the entity file of heartbeats of type FileType. Will prioritize
// local file if available.
func WithDetection(c Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			log.Debugln("execute dependency detection")

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

				language, ok := heartbeat.ParseLanguage(*h.Language)
				if !ok {
					log.Warnf("error parsing language of string %q", *h.Language)
				}

				dependencies, err := Detect(filepath, language)
				if err != nil {
					log.Warnf("error detecting dependencies of heartbeat: %s", err)
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
	var parser DependencyParser

	switch language {
	case heartbeat.LanguageC, heartbeat.LanguageCPP:
		parser = &ParserC{}
	case heartbeat.LanguageCSharp:
		parser = &ParserCSharp{}
	case heartbeat.LanguageElm:
		parser = &ParserElm{}
	case heartbeat.LanguageGo:
		parser = &ParserGo{}
	case heartbeat.LanguageHaskell:
		parser = &ParserHaskell{}
	case heartbeat.LanguageHaxe:
		parser = &ParserHaxe{}
	case heartbeat.LanguageHTML:
		parser = &ParserHTML{}
	case heartbeat.LanguageJava:
		parser = &ParserJava{}
	case heartbeat.LanguageJavaScript, heartbeat.LanguageTypeScript:
		parser = &ParserJavaScript{}
	case heartbeat.LanguageJSON:
		parser = &ParserJSON{}
	case heartbeat.LanguageKotlin:
		parser = &ParserKotlin{}
	case heartbeat.LanguageObjectiveC:
		parser = &ParserObjectiveC{}
	case heartbeat.LanguagePHP:
		parser = &ParserPHP{}
	case heartbeat.LanguagePython:
		parser = &ParserPython{}
	case heartbeat.LanguageRust:
		parser = &ParserRust{}
	case heartbeat.LanguageScala:
		parser = &ParserScala{}
	case heartbeat.LanguageSwift:
		parser = &ParserSwift{}
	case heartbeat.LanguageVBNet:
		parser = &ParserVbNet{}
	default:
		log.Debugf(
			"no parser has been found for language %q. Using Unknown parser to detect dependencies.", language)

		parser = &ParserUnknown{}
	}

	deps, err := parser.Parse(filepath)
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
		// filter max size
		if len(results) >= maxDependenciesCount {
			log.Debugf("max size of %d dependencies reached", maxDependenciesCount)
			break
		}

		// filter duplicate
		if _, ok := unique[d]; ok {
			continue
		}

		// filter dependencies off size
		if d == "" || len(d) > maxDependencyLength {
			log.Debugf(
				"dependency won't be sent because it's either empty or greater than %d characteres: %s",
				maxDependencyLength,
				d,
			)

			continue
		}

		unique[d] = struct{}{}

		results = append(results, d)
	}

	return results
}
