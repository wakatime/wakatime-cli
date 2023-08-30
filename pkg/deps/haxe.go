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

// Parse parses dependencies from Haxe file content using the chroma Haxe lexer.
func (p *ParserHaxe) Parse(filepath string) ([]string, error) {
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

	iter, err := lexers.Haxe.Tokenise(nil, string(data))
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
