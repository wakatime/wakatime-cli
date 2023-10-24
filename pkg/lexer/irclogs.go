package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// IRCLogs lexer.
type IRCLogs struct{}

// Lexer returns the lexer.
func (l IRCLogs) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"irc"},
			Filenames: []string{"*.weechatlog"},
			MimeTypes: []string{"text/x-irclog"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (IRCLogs) Name() string {
	return heartbeat.LanguageIRCLogs.StringChroma()
}
