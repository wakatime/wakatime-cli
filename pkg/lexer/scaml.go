package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Scaml lexer. For Scaml markup <http://scalate.fusesource.org/>. Scaml is Haml for Scala.
type Scaml struct{}

// Lexer returns the lexer.
func (l Scaml) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"scaml"},
			Filenames: []string{"*.scaml"},
			MimeTypes: []string{"text/x-scaml"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Scaml) Name() string {
	return heartbeat.LanguageScaml.StringChroma()
}
