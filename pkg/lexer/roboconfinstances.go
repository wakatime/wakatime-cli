package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// RoboconfInstances lexer for Roboconf <http://roboconf.net/en/roboconf.html> instances files.
type RoboconfInstances struct{}

// Lexer returns the lexer.
func (l RoboconfInstances) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"roboconf-instances"},
			Filenames: []string{"*.instances"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (RoboconfInstances) Name() string {
	return heartbeat.LanguageRoboconfInstances.StringChroma()
}
