package deps

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers/v"
)

var vbnetExcludeRegex = regexp.MustCompile(`(?i)^(system|microsoft)$`)

// StateVbNet is a token parsing state.
type StateVbNet int

const (
	// StateVbNetUnknown represents a unknown token parsing state.
	StateVbNetUnknown StateVbNet = iota
	// StateVbNetImport means we are in import section during token parsing.
	StateVbNetImport
)

// ParserVbNet is a dependency parser for the vb.net programming language.
// It is not thread safe.
type ParserVbNet struct {
	State  StateVbNet
	Buffer string
	Output []string
}

// Parse parses dependencies from VB.Net file content using the chroma VB.Net lexer.
func (p *ParserVbNet) Parse(filepath string) ([]string, error) {
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

	iter, err := v.VBNet.Tokenise(nil, string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize file content: %s", err)
	}

	for _, token := range iter.Tokens() {
		p.processToken(token)
	}

	return p.Output, nil
}

func (p *ParserVbNet) append(dep string) {
	dep = strings.TrimSpace(strings.Split(dep, ".")[0])

	if len(dep) == 0 {
		return
	}

	if vbnetExcludeRegex.MatchString(dep) {
		return
	}

	p.Output = append(p.Output, dep)
}

func (p *ParserVbNet) init() {
	p.Buffer = ""
	p.Output = nil
	p.State = StateVbNetUnknown
}

func (p *ParserVbNet) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.Keyword:
		p.processKeyword(token.Value)
	case chroma.Name, chroma.NameNamespace:
		p.processName(token.Value)
	case chroma.Operator:
		p.processOperator(token.Value)
	case chroma.Punctuation:
		p.processPunctuation(token.Value)
	case chroma.Text:
		p.processText(token.Value)
	}
}

func (p *ParserVbNet) processKeyword(value string) {
	if value == "Imports" {
		p.State = StateVbNetImport
		p.Buffer = ""
	}
}

func (p *ParserVbNet) processName(value string) {
	if p.State != StateVbNetImport {
		return
	}

	if value == "." {
		p.processPunctuation(value)
		return
	}

	p.Buffer += value
}

func (p *ParserVbNet) processPunctuation(value string) {
	if p.State != StateVbNetImport {
		return
	}

	if value == "." {
		p.Buffer += value
	}
}

func (p *ParserVbNet) processOperator(value string) {
	if p.State != StateVbNetImport {
		return
	}

	if value == "=" {
		p.Buffer = ""
	}
}

func (p *ParserVbNet) processText(value string) {
	if p.State != StateVbNetImport {
		return
	}

	if value == "\n" {
		p.append(p.Buffer)
		p.State = StateVbNetUnknown
		p.Buffer = ""
	}
}
