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

var pythonExcludeRegex = regexp.MustCompile(`(?i)^(os|sys|__[a-z]+__)$`)

// StatePython is a token parsing state.
type StatePython int

const (
	// StatePythonUnknown represents an unknown token parsing state.
	StatePythonUnknown StatePython = iota
	// StatePythonFrom means we are in from section of import during token parsing.
	StatePythonFrom
	// StatePythonImport means we are in import section during token parsing.
	StatePythonImport
)

// ParserPython is a dependency parser for the python programming language.
// It is not thread safe.
type ParserPython struct {
	State  StatePython
	Buffer string
	Output []string
}

// Parse parses dependencies from Python file content using the chroma Python lexer.
func (p *ParserPython) Parse(filepath string) ([]string, error) {
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

	l := lexers.Get(heartbeat.LanguagePython.String())
	if l == nil {
		return nil, fmt.Errorf("failed to get lexer for %s", heartbeat.LanguagePython.String())
	}

	iter, err := l.Tokenise(nil, string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize file content: %s", err)
	}

	t := iter.Tokens()

	for _, token := range t {
		p.processToken(token)
	}

	return p.Output, nil
}

func (p *ParserPython) append(dep string) {
	// if dot separated import path, select first element
	dep = strings.Split(dep, ".")[0]

	// trim whitespaces
	dep = strings.TrimSpace(dep)

	if len(dep) == 0 {
		return
	}

	// filter by regex
	if pythonExcludeRegex.MatchString(dep) {
		return
	}

	p.Output = append(p.Output, dep)
}

func (p *ParserPython) init() {
	p.State = StatePythonUnknown
	p.Output = []string{}
}

func (p *ParserPython) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.KeywordNamespace:
		p.processKeywordNamespace(token.Value)
	case chroma.Keyword:
		p.processKeyword(token.Value)
	case chroma.NameNamespace:
		p.processNameNamespace(token.Value)
	case chroma.Operator:
		p.processOperator(token.Value)
	case chroma.Text:
		p.processText(token.Value)
	}
}

func (p *ParserPython) processKeywordNamespace(value string) {
	switch value {
	case "from":
		p.State = StatePythonFrom
	case "import":
		p.State = StatePythonImport
	default:
		p.State = StatePythonUnknown
	}
}

func (p *ParserPython) processKeyword(value string) {
	if p.State == StatePythonImport && value == "as" {
		p.append(p.Buffer)
		p.Buffer = ""
		p.State = StatePythonUnknown
	}
}

func (p *ParserPython) processNameNamespace(value string) {
	switch p.State {
	case StatePythonFrom, StatePythonImport:
		p.Buffer += value
	default:
		p.State = StatePythonUnknown
	}
}

func (p *ParserPython) processOperator(value string) {
	if value != "," && p.State != StatePythonImport {
		return
	}

	p.append(p.Buffer)
	p.Buffer = ""
}

func (p *ParserPython) processText(value string) {
	if p.State != StatePythonImport && p.State != StatePythonFrom {
		return
	}

	if value == "\n" {
		p.append(p.Buffer)
		p.Buffer = ""
		p.State = StatePythonUnknown
	}
}
