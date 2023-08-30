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

var javaScriptExtensionRegex = regexp.MustCompile(`\.\w{1,4}$`)

// StateJavaScript is a token parsing state.
type StateJavaScript int

const (
	// StateJavaScriptUnknown represents an unknown token parsing state.
	StateJavaScriptUnknown StateJavaScript = iota
	// StateJavaScriptImport means we are in import section during token parsing.
	StateJavaScriptImport
)

// ParserJavaScript is a dependency parser for the JavaScript programming language.
// It is not thread safe.
type ParserJavaScript struct {
	State  StateJavaScript
	Output []string
}

// Parse parses dependencies from JavaScript file content using the chroma JavaScript lexer.
func (p *ParserJavaScript) Parse(filepath string) ([]string, error) {
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

	l := lexers.Get(heartbeat.LanguageJavaScript.String())
	if l == nil {
		return nil, fmt.Errorf("failed to get lexer for %s", heartbeat.LanguageJavaScript.String())
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

func (p *ParserJavaScript) append(dep string) {
	// trim whitespaces, single quotes and double quotes
	dep = strings.Trim(dep, `"' `)

	// if front slash path, select last element
	splitted := strings.Split(dep, `/`)
	dep = splitted[len(splitted)-1]

	// if back slash path, select last element
	splitted = strings.Split(dep, `\`)
	dep = splitted[len(splitted)-1]

	// remove extension
	dep = javaScriptExtensionRegex.ReplaceAllString(dep, "")

	p.Output = append(p.Output, dep)
}

func (p *ParserJavaScript) init() {
	p.State = StateJavaScriptUnknown
	p.Output = nil
}

func (p *ParserJavaScript) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.KeywordReserved:
		p.processKeywordReserved(token.Value)
	case chroma.LiteralStringSingle, chroma.LiteralStringDouble:
		p.processLiteralStringSingle(token.Value)
	case chroma.Punctuation:
		p.processPunctuation(token.Value)
	}
}

func (p *ParserJavaScript) processKeywordReserved(value string) {
	switch value {
	case "import":
		p.State = StateJavaScriptImport
	default:
		p.State = StateJavaScriptUnknown
	}
}

func (p *ParserJavaScript) processLiteralStringSingle(value string) {
	if p.State == StateJavaScriptImport {
		p.append(value)
	}

	p.State = StateJavaScriptUnknown
}

func (p *ParserJavaScript) processPunctuation(value string) {
	if p.State == StateJavaScriptImport && value == ";" {
		p.State = StateJavaScriptUnknown
	}
}
