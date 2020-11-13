package language

import (
	"fmt"
	"io"
	"os"
	fp "path/filepath"
	"sort"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	_ "github.com/alecthomas/chroma/lexers/a"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/b"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/c"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/circular" // not used directly
	_ "github.com/alecthomas/chroma/lexers/d"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/e"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/f"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/g"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/h"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/i"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/j"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/k"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/l"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/m"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/n"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/o"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/p"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/q"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/r"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/s"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/t"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/v"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/w"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/x"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/y"        // not used directly
	_ "github.com/alecthomas/chroma/lexers/z"        // not used directly
	"github.com/danwakefield/fnmatch"
	jww "github.com/spf13/jwalterweatherman"
)

// chromaMatchCustomized returns the best by filename matching lexer. Best lexer is determined
// by customized priority.
// This is a modified implementation of chroma.lexers.internal.api:Match().
func chromaMatchCustomized(filepath string) (heartbeat.Language, bool) {
	_, file := fp.Split(filepath)
	filename := fp.Base(file)
	matched := chroma.PrioritisedLexers{}

	// First, try primary filename matches.
	for _, lexer := range lexers.Registry.Lexers {
		config := lexer.Config()
		for _, glob := range config.Filenames {
			if fnmatch.Match(glob, filename, 0) {
				matched = append(matched, lexer)
			}
		}
	}

	if len(matched) > 0 {
		bestLexer := selectByCustomizedPriority(filepath, matched)

		language, ok := heartbeat.ParseLanguageFromChroma(bestLexer.Config().Name)
		if !ok {
			jww.WARN.Printf("failed to parse language from chroma lexer name %q", bestLexer.Config().Name)
			return heartbeat.LanguageUnknown, false
		}

		return language, true
	}

	// Next, try filename aliases.
	for _, lexer := range lexers.Registry.Lexers {
		config := lexer.Config()
		for _, glob := range config.AliasFilenames {
			if fnmatch.Match(glob, filename, 0) {
				matched = append(matched, lexer)
			}
		}
	}

	if len(matched) > 0 {
		bestLexer := selectByCustomizedPriority(filepath, matched)

		language, ok := heartbeat.ParseLanguageFromChroma(bestLexer.Config().Name)
		if !ok {
			jww.WARN.Printf("failed to parse language from chroma lexer name %q", bestLexer.Config().Name)
			return heartbeat.LanguageUnknown, false
		}

		return language, true
	}

	return heartbeat.LanguageUnknown, false
}

// weightedLexer is a lexer with priority and weight.
type weightedLexer struct {
	chroma.Lexer
	Weight   float32
	Priority float32
}

// selectByCustomizedPriority selects the best matching lexer by customized priority evaluation.
func selectByCustomizedPriority(filepath string, lexers chroma.PrioritisedLexers) chroma.Lexer {
	sort.Slice(lexers, func(i, j int) bool {
		icfg, jcfg := lexers[i].Config(), lexers[j].Config()

		// 1. by priority
		if icfg.Priority != jcfg.Priority {
			return icfg.Priority > jcfg.Priority
		}

		// 2. by name
		return strings.ToLower(icfg.Name) > strings.ToLower(jcfg.Name)
	})

	dir, _ := fp.Split(filepath)

	extensions, err := loadFolderExtensions(dir)
	if err != nil {
		jww.WARN.Printf("failed to load folder extensions: %s", err)
		return lexers[0]
	}

	head, err := fileHead(filepath)
	if err != nil {
		jww.WARN.Printf("failed to load head from file %q: %s", filepath, err)
		return lexers[0]
	}

	var weighted []weightedLexer

	for _, lexer := range lexers {
		var weight float32

		if analyser, ok := lexer.(chroma.Analyser); ok {
			weight = analyser.AnalyseText(string(head))
		}

		cfg := lexer.Config()

		if p, ok := priority(cfg.Name); ok {
			weighted = append(weighted, weightedLexer{
				Lexer:    lexer,
				Priority: p,
				Weight:   weight,
			})

			continue
		}

		if cfg.Name == "Matlab" {
			weighted = append(weighted, weightedLexer{
				Lexer:    lexer,
				Priority: cfg.Priority,
				Weight:   matlabWeight(weight, extensions),
			})

			continue
		}

		if cfg.Name == "Objective-C" {
			weighted = append(weighted, weightedLexer{
				Lexer:    lexer,
				Priority: cfg.Priority,
				Weight:   objectiveCWeight(weight, extensions),
			})

			continue
		}

		weighted = append(weighted, weightedLexer{
			Lexer:    lexer,
			Priority: cfg.Priority,
			Weight:   weight,
		})
	}

	sort.Slice(weighted, func(i, j int) bool {
		// 1. by weight
		if weighted[i].Weight != weighted[j].Weight {
			return weighted[i].Weight > weighted[j].Weight
		}

		// 2. by priority
		if weighted[i].Priority != weighted[j].Priority {
			return weighted[i].Priority > weighted[j].Priority
		}

		// 3. name
		return weighted[i].Lexer.Config().Name > weighted[j].Lexer.Config().Name
	})

	return weighted[0].Lexer
}

// fileHead returns the first 512000 bytes of the file's content.
func fileHead(filepath string) ([]byte, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s", err)
	}

	defer f.Close()

	data := make([]byte, 512000)

	_, err = f.ReadAt(data, 0)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read bytes from file: %s", err)
	}

	return data, nil
}

// objectiveCWeight determines the weight of objective-c by the provided same folder file extensions.
func objectiveCWeight(weight float32, extensions []string) float32 {
	var matFileExists bool

	for _, e := range extensions {
		if e == ".mat" {
			matFileExists = true
			break
		}
	}

	if matFileExists {
		weight -= 0.01
	} else {
		weight += 0.01
	}

	for _, e := range extensions {
		if e == ".h" {
			weight += 0.01
			break
		}
	}

	return weight
}

// matlabWeight determines the weight of matlab by the provided same folder file extensions.
func matlabWeight(weight float32, extensions []string) float32 {
	for _, e := range extensions {
		if e == ".mat" {
			weight += 0.01
			break
		}
	}

	var headerFileExists bool

	for _, e := range extensions {
		if e == ".h" {
			headerFileExists = true
			break
		}
	}

	if !headerFileExists {
		weight += 0.01
	}

	return weight
}

func chromaMatchOverwrite(filepath string) (heartbeat.Language, bool) {
	filepathLower := strings.ToLower(filepath)

	suffixes := chromaOverwriteTop()

	for suffix, language := range suffixes {
		if strings.HasSuffix(filepathLower, suffix) {
			return language, true
		}
	}

	return heartbeat.LanguageUnknown, false
}

func chromaOverwriteTop() map[string]heartbeat.Language {
	return map[string]heartbeat.Language{
		"/cmakelists.txt": heartbeat.LanguageCMake,
		"/go.mod":         heartbeat.LanguageGo,
	}
}
