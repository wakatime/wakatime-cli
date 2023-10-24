package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// EarlGrey lexer.
type EarlGrey struct{}

// Lexer returns the lexer.
func (l EarlGrey) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"earl-grey", "earlgrey", "eg"},
			Filenames: []string{"*.eg"},
			MimeTypes: []string{"text/x-earl-grey"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (EarlGrey) Name() string {
	return heartbeat.LanguageEarlGrey.StringChroma()
}
