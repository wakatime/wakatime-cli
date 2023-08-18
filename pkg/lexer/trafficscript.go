package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// TrafficScript lexer. For `Riverbed Stingray Traffic Manager
// <http://www.riverbed.com/stingray>`
type TrafficScript struct{}

// Lexer returns the lexer.
func (l TrafficScript) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"rts", "trafficscript"},
			Filenames: []string{"*.rts"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (TrafficScript) Name() string {
	return heartbeat.LanguageTrafficScript.StringChroma()
}
