package deps

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/alecthomas/chroma"
)

// StateHaskell is a token parsing state.
type StateHaskell int

const (
	// StateHaskellUnknown represents an unknown token parsing state.
	StateHaskellUnknown StateHaskell = iota
	// StateHaskellImport means we are in import section during token parsing.
	StateHaskellImport
)

// ParserHaskell is a dependency parser for the Haskell programming language.
// It is not thread safe.
type ParserHaskell struct {
	State  StateHaskell
	Output []string
}

// Parse parses dependencies from Haskell file content via ReadCloser using the chroma Haskell lexer.
func (p *ParserHaskell) Parse(reader io.ReadCloser, lexer chroma.Lexer) ([]string, error) {
	defer reader.Close()

	p.init()
	defer p.init()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %s", err)
	}

	iter, err := lexer.Tokenise(nil, string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize file content: %s", err)
	}

	for _, token := range iter.Tokens() {
		p.processToken(token)
	}

	return p.Output, nil
}

func (p *ParserHaskell) append(dep string) {
	// trim whitespaces
	dep = strings.TrimSpace(dep)

	// if dot separated import path, select first element
	dep = strings.Split(dep, ".")[0]

	// trim whitespaces
	dep = strings.TrimSpace(dep)

	p.Output = append(p.Output, dep)
}

func (p *ParserHaskell) init() {
	p.State = StateHaskellUnknown
	p.Output = []string{}
}

func (p *ParserHaskell) processToken(token chroma.Token) {
	switch {
	case token.Type == chroma.KeywordReserved:
		p.processKeywordReserved(token.Value)
	case token.Type == chroma.Keyword:
		p.processKeyword(token.Value)
	case token.Type == chroma.NameNamespace:
		p.processNameNamespace(token.Value)
	case token.Type != chroma.Text:
		p.State = StateHaskellUnknown
	}
}

func (p *ParserHaskell) processKeywordReserved(value string) {
	switch strings.TrimSpace(value) {
	case "import":
		p.State = StateHaskellImport
	default:
		p.State = StateHaskellUnknown
	}
}

func (p *ParserHaskell) processKeyword(value string) {
	if p.State != StateHaskellImport || strings.TrimSpace(value) != "qualified" {
		p.State = StateHaskellUnknown
	}
}

func (p *ParserHaskell) processNameNamespace(value string) {
	if p.State == StateHaskellImport {
		p.append(value)
	}
}
