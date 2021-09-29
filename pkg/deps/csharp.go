package deps

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers/c"
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

// Parse parses dependencies from C# file content using the chroma C# lexer.
func (p *ParserCSharp) Parse(filepath string) ([]string, error) {
	reader, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %s", filepath, err)
	}

	defer reader.Close()

	p.init()
	defer p.init()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %s", err)
	}

	iter, err := c.CSharp.Tokenise(nil, string(data))
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
