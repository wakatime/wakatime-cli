package deps

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/alecthomas/chroma"
)

// StateElm is a token parsing state.
type StateElm int

const (
	// StateElmUnknown represents a unknown token parsing state.
	StateElmUnknown StateElm = iota
	// StateElmImport means we are in import during token parsing.
	StateElmImport
)

// ParserElm is a dependency parser for the elm programming language.
// It is not thread safe.
type ParserElm struct {
	State  StateElm
	Output []string
}

// Parse parses dependencies from elm file content via ReadCloser using the chroma elm lexer.
func (p *ParserElm) Parse(reader io.ReadCloser, lexer chroma.Lexer) ([]string, error) {
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

func (p *ParserElm) append(dep string) {
	p.Output = append(p.Output, strings.TrimSpace(strings.Split(dep, ".")[0]))
}

func (p *ParserElm) init() {
	p.Output = []string{}
}

func (p *ParserElm) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.KeywordNamespace:
		p.processKeywordNamespace(token.Value)
	case chroma.NameClass:
		p.processNameClass(token.Value)
	default:
		p.State = StateElmUnknown
	}
}

func (p *ParserElm) processKeywordNamespace(value string) {
	if strings.TrimSpace(value) == "import" {
		p.State = StateElmImport
	} else {
		p.State = StateElmUnknown
	}
}

func (p *ParserElm) processNameClass(value string) {
	if p.State == StateElmImport {
		p.append(value)
	}
}
