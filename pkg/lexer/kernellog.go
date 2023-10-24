package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// KernelLog lexer.
type KernelLog struct{}

// Lexer returns the lexer.
func (l KernelLog) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"kmsg", "dmesg"},
			Filenames: []string{"*.kmsg", "*.dmesg"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (KernelLog) Name() string {
	return heartbeat.LanguageKernelLog.StringChroma()
}
