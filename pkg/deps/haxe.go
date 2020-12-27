package deps

import (
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma"
)

// nolint:noglobal
var haxeExcludeRegex = regexp.MustCompile(`(?i)^haxe$`)

// StateHaxe is a token parsing state.
type StateHaxe int

const (
	// StateHaxeUnknown represents an unknown token parsing state.
	StateHaxeUnknown StateHaxe = iota
	// StateHaxeImport means we are in import section during token parsing.
	StateHaxeImport
)

// ParserHaxe is a dependency parser for the Haxe programming language.
// It is not thread safe.
type ParserHaxe struct {
	State  StateHaxe
	Output []string
}

// Parse parses dependencies from Haxe file content via ReadCloser using the chroma Haxe lexer.
func (p *ParserHaxe) Parse(reader io.ReadCloser, lexer chroma.Lexer) ([]string, error) {
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

func (p *ParserHaxe) append(dep string) {
	dep = strings.TrimSpace(dep)

	if haxeExcludeRegex.MatchString(dep) {
		return
	}

	p.Output = append(p.Output, dep)
}

func (p *ParserHaxe) init() {
	p.State = StateHaxeUnknown
	p.Output = []string{}
}

func (p *ParserHaxe) processToken(token chroma.Token) {
	switch {
	case token.Type == chroma.KeywordNamespace:
		p.processKeywordNamespace(token.Value)
	case token.Type == chroma.NameNamespace:
		p.processNameNamespace(token.Value)
	case token.Type != chroma.Text:
		p.State = StateHaxeUnknown
	}
}

func (p *ParserHaxe) processKeywordNamespace(value string) {
	switch value {
	case "import":
		p.State = StateHaxeImport
	default:
		p.State = StateHaxeUnknown
	}
}

func (p *ParserHaxe) processNameNamespace(value string) {
	if p.State == StateHaxeImport {
		p.append(value)
	}

	p.State = StateHaxeUnknown
}
