package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// Slurm lexer. Lexer for (ba|k|z|)sh Slurm scripts.
type Slurm struct{}

// Lexer returns the lexer.
func (l Slurm) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"slurm", "sbatch"},
			Filenames: []string{"*.sl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		bash := lexers.Get(heartbeat.LanguageBash.StringChroma())
		if bash == nil {
			return 0
		}

		return bash.AnalyseText(text)
	})

	return lexer
}

// Name returns the name of the lexer.
func (Slurm) Name() string {
	return heartbeat.LanguageSlurm.StringChroma()
}
