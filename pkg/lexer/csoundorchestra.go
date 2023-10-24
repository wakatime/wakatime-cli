package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// CsoundOrchestra lexer.
type CsoundOrchestra struct{}

// Lexer returns the lexer.
func (l CsoundOrchestra) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"csound", "csound-orc"},
			Filenames: []string{"*.orc", "*.udo"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (CsoundOrchestra) Name() string {
	return heartbeat.LanguageCsoundOrchestra.StringChroma()
}
