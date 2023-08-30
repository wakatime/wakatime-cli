package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// RoboconfGraph lexer for Roboconf <http://roboconf.net/en/roboconf.html> graph files.
type RoboconfGraph struct{}

// Lexer returns the lexer.
func (l RoboconfGraph) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"roboconf-graph"},
			Filenames: []string{"*.graph"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (RoboconfGraph) Name() string {
	return heartbeat.LanguageRoboconfGraph.StringChroma()
}
