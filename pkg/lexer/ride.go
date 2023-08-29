package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Ride lexer. For Ride <https://docs.wavesplatform.com/en/ride/about-ride.html>
// source code.
type Ride struct{}

// Lexer returns the lexer.
func (l Ride) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"ride"},
			Filenames: []string{"*.ride"},
			MimeTypes: []string{"text/x-ride"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Ride) Name() string {
	return heartbeat.LanguageRide.StringChroma()
}
