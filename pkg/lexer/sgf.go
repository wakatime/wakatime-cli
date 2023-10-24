package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// SmartGameFormat lexer. Lexer for Smart Game Format (sgf) file format.
//
// The format is used to store game records of board games for two players
// (mainly Go game). For more information about the definition of the format,
// see: https://www.red-bean.com/sgf/
type SmartGameFormat struct{}

// Lexer returns the lexer.
func (l SmartGameFormat) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"sgf"},
			Filenames: []string{"*.sgf"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (SmartGameFormat) Name() string {
	return heartbeat.LanguageSmartGameFormat.StringChroma()
}
