package deps

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var kotlinExcludeRegex = regexp.MustCompile(`(?i)^java\.`)

// StateKotlin is a token parsing state.
type StateKotlin int

const (
	// StateKotlinUnknown represents an unknown token parsing state.
	StateKotlinUnknown StateKotlin = iota
	// StateKotlinImport means we are in import section during token parsing.
	StateKotlinImport
)

// ParserKotlin is a dependency parser for the Kotlin programming language.
// It is not thread safe.
type ParserKotlin struct {
	State  StateKotlin
	Output []string
}

// Parse parses dependencies from Kotlin file content using the chroma Kotlin lexer.
func (p *ParserKotlin) Parse(filepath string) ([]string, error) {
	reader, err := os.Open(filepath) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %s", filepath, err)
	}

	defer func() {
		if err := reader.Close(); err != nil {
			log.Debugf("failed to close file: %s", err)
		}
	}()

	p.init()
	defer p.init()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %s", err)
	}

	l := lexers.Get(heartbeat.LanguageKotlin.String())
	if l == nil {
		return nil, fmt.Errorf("failed to get lexer for %s", heartbeat.LanguageKotlin.String())
	}

	iter, err := l.Tokenise(nil, string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize file content: %s", err)
	}

	for _, token := range iter.Tokens() {
		p.processToken(token)
	}

	return p.Output, nil
}

func (p *ParserKotlin) append(dep string) {
	splitted := strings.Split(dep, ".")

	// remove trailing asterisk, if existing
	if splitted[len(splitted)-1] == "*" {
		splitted = splitted[:len(splitted)-1]
	}

	if len(splitted) == 0 {
		return
	}

	if len(splitted) == 1 {
		dep = splitted[0]
	} else {
		// only consider the first two elements
		dep = strings.Join(splitted[:2], ".")
	}

	// trim whitespace
	dep = strings.TrimSpace(dep)

	// filter by exclude regex
	if kotlinExcludeRegex.MatchString(dep) {
		return
	}

	p.Output = append(p.Output, dep)
}

func (p *ParserKotlin) init() {
	p.State = StateKotlinUnknown
	p.Output = nil
}

func (p *ParserKotlin) processToken(token chroma.Token) {
	switch {
	case token.Type == chroma.Keyword:
		p.processKeyword(token.Value)
	case token.Type == chroma.NameNamespace:
		p.processNameNamespace(token.Value)
	case token.Type != chroma.Text:
		p.State = StateKotlinUnknown
	}
}

func (p *ParserKotlin) processKeyword(value string) {
	switch value {
	case "import":
		p.State = StateKotlinImport
	default:
		p.State = StateKotlinUnknown
	}
}

func (p *ParserKotlin) processNameNamespace(value string) {
	switch p.State {
	case StateKotlinImport:
		p.append(value)
	default:
		p.State = StateKotlinUnknown
	}
}
