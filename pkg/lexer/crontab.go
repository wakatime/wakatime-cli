package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Crontab lexer.
type Crontab struct{}

// Lexer returns the lexer.
func (l Crontab) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"crontab"},
			Filenames: []string{"crontab"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Crontab) Name() string {
	return heartbeat.LanguageCrontab.StringChroma()
}
