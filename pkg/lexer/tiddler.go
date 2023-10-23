package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Tiddler lexer. For TiddlyWiki5 <https://tiddlywiki.com/#TiddlerFiles> markup.
type Tiddler struct{}

// Lexer returns the lexer.
func (l Tiddler) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"tid"},
			Filenames: []string{"*.tid"},
			MimeTypes: []string{"text/vnd.tiddlywiki"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Tiddler) Name() string {
	return heartbeat.LanguageTiddler.StringChroma()
}
