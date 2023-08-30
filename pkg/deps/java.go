package deps

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var javaExcludeRegex = regexp.MustCompile(`(?i)^(java\..*|javax\..*)`)

// StateJava is a token parsing state.
type StateJava int

const (
	// StateJavaUnknown represents a unknown token parsing state.
	StateJavaUnknown StateJava = iota
	// StateJavaImport means we are in import section during token parsing.
	StateJavaImport
	// StateJavaImportFinished means we finished import section during token parsing.
	StateJavaImportFinished
)

// ParserJava is a dependency parser for the java programming language.
// It is not thread safe.
type ParserJava struct {
	Buffer string
	Output []string
	State  StateJava
}

// Parse parses dependencies from Java file content using the chroma Java lexer.
func (p *ParserJava) Parse(filepath string) ([]string, error) {
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

	l := lexers.Get(heartbeat.LanguageJava.String())
	if l == nil {
		return nil, fmt.Errorf("failed to get lexer for %s", heartbeat.LanguageJava.String())
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

func (p *ParserJava) append(dep string) {
	if javaExcludeRegex.MatchString(dep) {
		return
	}

	if len(strings.TrimSpace(dep)) == 0 {
		return
	}

	p.Output = append(p.Output, dep)
}

func (p *ParserJava) init() {
	p.Buffer = ""
	p.State = StateJavaUnknown
	p.Output = nil
}

func (p *ParserJava) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.KeywordNamespace:
		p.processKeywordNamespace(token.Value)
	case chroma.Name:
		p.processName(token.Value)
	case chroma.NameAttribute:
		p.processNameAttribute(token.Value)
	case chroma.NameNamespace:
		p.processNameNamespace(token.Value)
	case chroma.Operator:
		p.processOperator(token.Value)
	}
}

func (p *ParserJava) processKeywordNamespace(value string) {
	splitted := strings.Fields(value)
	if len(splitted) > 0 && splitted[0] == "import" {
		p.State = StateJavaImport
		return
	}

	if p.State != StateJavaImportFinished {
		return
	}

	splitted = strings.Split(value, ".")
	if len(splitted) == 1 {
		p.append(splitted[0])
		p.State = StateJavaUnknown

		return
	}

	if len(splitted) == 0 {
		p.State = StateJavaUnknown
		return
	}

	// remove leading top-level domain
	if len(splitted[0]) == 3 {
		splitted = splitted[1:]
	}

	// remove trailing asterisk
	if splitted[len(splitted)-1] == "*" {
		splitted = splitted[:len(splitted)-1]
	}

	switch {
	case len(splitted) == 1:
		p.append(splitted[0])
	case len(splitted) > 1:
		// use first 2 elements
		p.append(strings.Join(splitted[:2], "."))
	}

	p.State = StateJavaUnknown
}

func (p *ParserJava) processName(value string) {
	if p.State == StateJavaImport {
		p.Buffer += value
	}
}

func (p *ParserJava) processNameAttribute(value string) {
	if p.State == StateJavaImport {
		p.Buffer += value
	}
}

func (p *ParserJava) processNameNamespace(value string) {
	if p.State == StateJavaImport && value != "package" && value != "namespace" && value != "static" {
		p.Buffer += value
	}
}

func (p *ParserJava) processOperator(value string) {
	if value == ";" {
		p.State = StateJavaImportFinished
		p.processKeywordNamespace(p.Buffer)
		p.State = StateJavaUnknown
		p.Buffer = ""

		return
	}

	if p.State == StateJavaImport {
		p.Buffer += value
	}
}
