package deps

import (
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma"
)

var goExcludeRegex = regexp.MustCompile(`(?i)^"fmt"$`)

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

// Parse parses dependencies from golang file content via ReadCloser using the chroma golang lexer.
func (p *ParserGo) Parse(reader io.ReadCloser, lexer chroma.Lexer) ([]string, error) {
	defer reader.Close()

	p.init()
	defer p.init()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %s", err)
	}

	iter, err := lexer.Tokenise(&chroma.TokeniseOptions{
		State:    "root",
		EnsureLF: true,
	}, string(data))
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
	p.Output = nil
	p.Parenthesis = 0
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
