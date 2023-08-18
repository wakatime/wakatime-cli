package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Sqlite3con lexer. Lexer for example sessions using sqlite3.
type Sqlite3con struct{}

// Lexer returns the lexer.
func (l Sqlite3con) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"sqlite3"},
			Filenames: []string{"*.sqlite3-console"},
			MimeTypes: []string{"text/x-sqlite3-console"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Sqlite3con) Name() string {
	return heartbeat.LanguageSqlite3con.StringChroma()
}
