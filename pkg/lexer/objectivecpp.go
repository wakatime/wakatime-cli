package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// ObjectiveCPP lexer.
type ObjectiveCPP struct{}

// Lexer returns the lexer.
func (l ObjectiveCPP) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"objective-c++", "objectivec++", "obj-c++", "objc++"},
			Filenames: []string{"*.mm", "*.hh"},
			MimeTypes: []string{"text/x-objective-c++"},
			// Lower than C++.
			Priority: 0.05,
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (ObjectiveCPP) Name() string {
	return heartbeat.LanguageObjectiveCPP.StringChroma()
}
