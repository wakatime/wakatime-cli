package deps

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers/r"
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

// Parse parses dependencies from Rust file content using the chroma Rust lexer.
func (p *ParserRust) Parse(filepath string) ([]string, error) {
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

	iter, err := r.Rust.Tokenise(nil, string(data))
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
