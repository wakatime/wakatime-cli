package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// PostgresConsole lexer.
type PostgresConsole struct{}

// Lexer returns the lexer.
func (l PostgresConsole) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"psql", "postgresql-console", "postgres-console"},
			MimeTypes: []string{"text/x-postgresql-psql"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (PostgresConsole) Name() string {
	return heartbeat.LanguagePostgresConsole.StringChroma()
}
