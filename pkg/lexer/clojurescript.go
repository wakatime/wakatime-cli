package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// ClojureScript lexer.
type ClojureScript struct{}

// Lexer returns the lexer.
func (l ClojureScript) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"clojurescript", "cljs"},
			Filenames: []string{"*.cljs"},
			MimeTypes: []string{"text/x-clojurescript", "application/x-clojurescript"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (ClojureScript) Name() string {
	return heartbeat.LanguageClojureScript.StringChroma()
}
