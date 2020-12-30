package deps

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma"
	lp "github.com/alecthomas/chroma/lexers/p"
)

// nolint:noglobal
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
	Parenthesis int
	State       StatePython
	Output      []string
}

// Parse parses dependencies from Python file content using the chroma Python lexer.
func (p *ParserPython) Parse(filepath string) ([]string, error) {
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

	iter, err := lp.Python.Tokenise(nil, string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize file content: %s", err)
	}

	for _, token := range iter.Tokens() {
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
	p.Parenthesis = 0
	p.State = StatePythonUnknown
	p.Output = []string{}
}

func (p *ParserPython) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.KeywordNamespace:
		p.processKeywordNamespace(token.Value)
	case chroma.NameNamespace:
		p.processNameNamespace(token.Value)
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

func (p *ParserPython) processNameNamespace(value string) {
	switch p.State {
	case StatePythonFrom:
		p.append(value)
	case StatePythonImport:
		p.append(value)
	default:
		p.State = StatePythonUnknown
	}
}
