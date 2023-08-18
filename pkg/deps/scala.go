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

// StateScala is a token parsing state.
type StateScala int

const (
	// StateScalaUnknown represents an unknown token parsing state.
	StateScalaUnknown StateScala = iota
	// StateScalaImport means we are in import section during token parsing.
	StateScalaImport
)

// ParserScala is a dependency parser for the Scala programming language.
// It is not thread safe.
type ParserScala struct {
	State  StateScala
	Output []string
}

// Parse parses dependencies from Scala file content using the chroma Scala lexer.
func (p *ParserScala) Parse(filepath string) ([]string, error) {
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

	l := lexers.Get(heartbeat.LanguageScala.String())
	if l == nil {
		return nil, fmt.Errorf("failed to get lexer for %s", heartbeat.LanguageScala.String())
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

func (p *ParserScala) append(dep string) {
	dep = strings.TrimSpace(dep)
	dep = strings.TrimPrefix(dep, "__root__")
	dep = strings.Trim(dep, "_. ")

	p.Output = append(p.Output, dep)
}

func (p *ParserScala) init() {
	p.State = StateScalaUnknown
	p.Output = nil
}

func (p *ParserScala) processToken(token chroma.Token) {
	switch {
	case token.Type == chroma.Keyword:
		p.processKeyword(token.Value)
	case token.Type == chroma.NameNamespace:
		p.processNameNamespace(token.Value)
	case token.Type != chroma.Text:
		p.State = StateScalaUnknown
	}
}

func (p *ParserScala) processKeyword(value string) {
	switch value {
	case "import":
		p.State = StateScalaImport
	default:
		p.State = StateScalaUnknown
	}
}

func (p *ParserScala) processNameNamespace(value string) {
	switch p.State {
	case StateScalaImport:
		p.append(value)
	default:
		p.State = StateScalaUnknown
	}
}
