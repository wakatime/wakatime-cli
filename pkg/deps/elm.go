package deps

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers/e"
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

// Parse parses dependencies from Elm file content using the chroma Elm lexer.
func (p *ParserElm) Parse(filepath string) ([]string, error) {
	reader, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %s", filepath, err)
	}

	defer reader.Close()

	p.init()
	defer p.init()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %s", err)
	}

	iter, err := e.Elm.Tokenise(nil, string(data))
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
	p.State = StateElmUnknown
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
