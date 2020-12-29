package deps

import (
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma"
)

var csharpExcludeRegex = regexp.MustCompile(`(?i)^(system|microsoft)$`)

// StateCSharp is a token parsing state.
type StateCSharp int

const (
	// StateCSharpUnknown represents a unknown token parsing state.
	StateCSharpUnknown StateCSharp = iota
	// StateCSharpImport means we are in import section during token parsing.
	StateCSharpImport
)

// ParserCSharp is a dependency parser for the c# programming language.
// It is not thread safe.
type ParserCSharp struct {
	State  StateCSharp
	Buffer string
	Output []string
}

// Parse parses dependencies from c# file content via ReadCloser using the chroma c# lexer.
func (p *ParserCSharp) Parse(reader io.ReadCloser, lexer chroma.Lexer) ([]string, error) {
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

func (p *ParserCSharp) append(dep string) {
	dep = strings.TrimSpace(strings.Split(dep, ".")[0])

	if len(dep) == 0 {
		return
	}

	if csharpExcludeRegex.MatchString(dep) {
		return
	}

	p.Output = append(p.Output, dep)
}

func (p *ParserCSharp) init() {
	p.State = StateCSharpUnknown
	p.Output = nil
}

func (p *ParserCSharp) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.Keyword:
		p.processKeyword(token.Value)
	case chroma.Name, chroma.NameNamespace:
		p.processName(token.Value)
	case chroma.Punctuation:
		p.processPunctuation(token.Value)
	}
}

func (p *ParserCSharp) processKeyword(value string) {
	if value == "using" {
		p.State = StateCSharpImport
		p.Buffer = ""
	}
}

func (p *ParserCSharp) processName(value string) {
	if p.State != StateCSharpImport {
		return
	}

	switch value {
	case "import", "package", "namespace", "static":
	default:
		p.Buffer += value
	}
}

func (p *ParserCSharp) processPunctuation(value string) {
	if p.State != StateCSharpImport {
		return
	}

	switch value {
	case ";":
		p.append(p.Buffer)
		p.Buffer = ""
		p.State = StateCSharpUnknown
	case "=":
		p.Buffer = ""
	default:
		p.Buffer += value
	}
}
