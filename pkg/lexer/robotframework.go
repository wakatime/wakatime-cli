package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// RobotFramework lexer for Robot Framework <http://robotframework.org> test data.
type RobotFramework struct{}

// Lexer returns the lexer.
func (l RobotFramework) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"robotframework"},
			Filenames: []string{"*.robot"},
			MimeTypes: []string{"text/x-robotframework"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (RobotFramework) Name() string {
	return heartbeat.LanguageRobotFramework.StringChroma()
}
