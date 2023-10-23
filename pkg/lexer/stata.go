package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Stata lexer. For Stata <http://www.stata.com/> do files.
//
// Syntax based on
// - http://fmwww.bc.edu/RePEc/bocode/s/synlightlist.ado
// - https://github.com/isagalaev/highlight.js/blob/master/src/languages/stata.js
// - https://github.com/jpitblado/vim-stata/blob/master/syntax/stata.vim
type Stata struct{}

// Lexer returns the lexer.
func (l Stata) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"stata", "do"},
			Filenames: []string{"*.do", "*.ado"},
			MimeTypes: []string{"text/x-stata", "text/stata", "application/x-stata"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Stata) Name() string {
	return heartbeat.LanguageStata.StringChroma()
}
