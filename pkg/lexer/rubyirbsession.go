package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// RubyIRBSession lexer. For Ruby interactive console (**irb**) output.
type RubyIRBSession struct{}

// Lexer returns the lexer.
func (l RubyIRBSession) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"rbcon", "irb"},
			MimeTypes: []string{"text/x-ruby-shellsession"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (RubyIRBSession) Name() string {
	return heartbeat.LanguageRubyIRBSession.StringChroma()
}
