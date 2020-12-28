package deps

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/alecthomas/chroma"
)

// StateRust is a token parsing state.
type StateRust int

const (
	// StateRustUnknown represents a unknown token parsing state.
	StateRustUnknown StateRust = iota
	// StateRustExtern means we are in extern section during token parsing.
	StateRustExtern
	// StateRustExternCrate means we are in extern crate section during token parsing.
	StateRustExternCrate
)

// ParserRust is a dependency parser for the rust programming language.
// It is not thread safe.
type ParserRust struct {
	State  StateRust
	Output []string
}

// Parse parses dependencies from rust file content via ReadCloser using the chroma rust lexer.
func (p *ParserRust) Parse(reader io.ReadCloser, lexer chroma.Lexer) ([]string, error) {
	defer reader.Close()

	p.init()
	defer p.init()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %s", err)
	}

	iter, err := lexer.Tokenise(&chroma.TokeniseOptions{
		State:    "root",
		EnsureLF: true,
	}, string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize file content: %s", err)
	}

	for _, token := range iter.Tokens() {
		p.processToken(token)
	}

	return p.Output, nil
}

func (p *ParserRust) init() {
	p.State = StateRustUnknown
	p.Output = []string{}
}

func (p *ParserRust) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.Keyword:
		p.processKeyword(token.Value)
	case chroma.Name:
		p.processName(token.Value)
	}
}

func (p *ParserRust) processKeyword(value string) {
	if p.State == StateRustExtern && value == "crate" {
		p.State = StateRustExternCrate
	} else if value == "extern" {
		p.State = StateRustExtern
	}
}

func (p *ParserRust) processName(value string) {
	if p.State == StateRustExternCrate {
		p.Output = append(p.Output, strings.TrimSpace(value))
	}

	p.State = StateRustUnknown
}
