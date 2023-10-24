package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Todotxt lexer. Lexer for Todo.txt <http://todotxt.com/> todo list format.
type Todotxt struct{}

// Lexer returns the lexer.
func (l Todotxt) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:    l.Name(),
			Aliases: []string{"todotxt"},
			// *.todotxt is not a standard extension for Todo.txt files; including it
			// makes testing easier, and also makes autodetecting file type easier.
			Filenames: []string{"todo.txt", "*.todotxt"},
			MimeTypes: []string{"text/x-todo"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Todotxt) Name() string {
	return heartbeat.LanguageTodotxt.StringChroma()
}
