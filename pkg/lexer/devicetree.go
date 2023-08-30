package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Devicetree lexer.
type Devicetree struct{}

// Lexer returns the lexer.
func (l Devicetree) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"devicetree", "dts"},
			Filenames: []string{"*.dts", "*.dtsi"},
			MimeTypes: []string{"text/x-c"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Devicetree) Name() string {
	return heartbeat.LanguageDevicetree.StringChroma()
}
