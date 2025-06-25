package deps

import (
	"context"
	"fmt"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/file"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
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
func (p *ParserRust) Parse(ctx context.Context, filepath string) ([]string, error) {
	head, err := file.ReadHead(ctx, filepath, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read: %s", err)
	}

	p.init()
	defer p.init()

	l := lexers.Get(heartbeat.LanguageRust.String())
	if l == nil {
		return nil, fmt.Errorf("failed to get lexer for %s", heartbeat.LanguageRust.String())
	}

	iter, err := l.Tokenise(nil, string(head))
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
