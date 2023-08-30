package deps

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
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

// Parse parses dependencies from Haskell file content using the chroma Haskell lexer.
func (p *ParserHaskell) Parse(filepath string) ([]string, error) {
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

	l := lexers.Get(heartbeat.LanguageHaskell.String())
	if l == nil {
		return nil, fmt.Errorf("failed to get lexer for %s", heartbeat.LanguageHaskell.String())
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
