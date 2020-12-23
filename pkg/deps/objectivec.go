package deps

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/alecthomas/chroma"
)

// StateObjectiveC is a token parsing state.
type StateObjectiveC int

const (
	// StateObjectiveCUnknown represents a unknown token parsing state.
	StateObjectiveCUnknown StateObjectiveC = iota
	// StateObjectiveCHash means we are in hash section during token parsing.
	StateObjectiveCHash
)

// ParserObjectiveC is a dependency parser for the objective-c programming language.
// It is not thread safe.
type ParserObjectiveC struct {
	State  StateObjectiveC
	Output []string
}

// Parse parses dependencies from objective-c file content via ReadCloser using the chroma objective-c lexer.
func (p *ParserObjectiveC) Parse(reader io.ReadCloser, lexer chroma.Lexer) ([]string, error) {
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

func (p *ParserObjectiveC) append(dep string) {
	// remove import prefix
	dep = strings.TrimPrefix(dep, "import ")
	dep = strings.TrimSpace(dep)

	// remove surrounding quotes
	dep = strings.Trim(dep, `"' `)

	// remove surrounding angle brackets
	dep = strings.TrimLeft(dep, "< ")
	dep = strings.TrimRight(dep, "> ")

	// only consider first part of an import path
	dep = strings.Split(dep, "/")[0]

	// trim extension
	dep = strings.TrimSuffix(dep, ".h")
	dep = strings.TrimSuffix(dep, ".m")

	p.Output = append(p.Output, strings.TrimSpace(dep))
}

func (p *ParserObjectiveC) init() {
	p.State = StateObjectiveCUnknown
	p.Output = nil
}

func (p *ParserObjectiveC) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.CommentPreproc:
		p.processCommentPreproc(token.Value)
	default:
		p.State = StateObjectiveCUnknown
	}
}

func (p *ParserObjectiveC) processCommentPreproc(value string) {
	switch {
	case value == "#":
		p.State = StateObjectiveCHash
	case p.State == StateObjectiveCHash && strings.HasPrefix(value, "import "):
		p.append(value)
		p.State = StateObjectiveCUnknown
	default:
		p.State = StateObjectiveCUnknown
	}
}
