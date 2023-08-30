package deps

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var goExcludeRegex = regexp.MustCompile(`^"fmt"$`)

// StateGo is a token parsing state.
type StateGo int

const (
	// StateGoUnknown represents a unknown token parsing state.
	StateGoUnknown StateGo = iota
	// StateGoImport means we are in import section during token parsing.
	StateGoImport
)

// ParserGo is a dependency parser for the go programming language.
// It is not thread safe.
type ParserGo struct {
	Parenthesis int
	State       StateGo
	Output      []string
}

// Parse parses dependencies from Golang file content using the chroma Golang lexer.
func (p *ParserGo) Parse(filepath string) ([]string, error) {
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

	iter, err := lexers.Go.Tokenise(nil, string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize file content: %s", err)
	}

	for _, token := range iter.Tokens() {
		p.processToken(token)
	}

	return p.Output, nil
}

func (p *ParserGo) append(dep string) {
	if goExcludeRegex.MatchString(dep) {
		return
	}

	p.Output = append(p.Output, strings.Trim(dep, `" `))
}

func (p *ParserGo) init() {
	p.Parenthesis = 0
	p.State = StateGoUnknown
	p.Output = nil
}

func (p *ParserGo) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.KeywordNamespace:
		p.processKeywordNamespace(token.Value)
	case chroma.Punctuation:
		p.processPunctuation(token.Value)
	case chroma.LiteralString:
		p.processLiteralString(token.Value)
	case chroma.Text:
		p.processText(token.Value)
	}
}

func (p *ParserGo) processKeywordNamespace(value string) {
	p.Parenthesis = 0

	switch value {
	case "import":
		p.State = StateGoImport
	default:
		p.State = StateGoUnknown
	}
}

func (p *ParserGo) processPunctuation(value string) {
	switch value {
	case "(":
		p.Parenthesis++
	case ")":
		p.Parenthesis--
	}
}

func (p *ParserGo) processLiteralString(value string) {
	if p.State == StateGoImport {
		p.append(value)
	}
}

func (p *ParserGo) processText(value string) {
	if p.State == StateGoImport {
		if value == "\n" && p.Parenthesis <= 0 {
			p.State = StateGoUnknown
			p.Parenthesis = 0
		}
	} else {
		p.State = StateGoUnknown
	}
}
