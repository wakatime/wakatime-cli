package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// POVRay lexer.
type POVRay struct{}

// Lexer returns the lexer.
func (l POVRay) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			// POVRAY is similar to JSON/C, but the combination of camera and
			// light_source is probably not very likely elsewhere. HLSL or GLSL
			// are similar (GLSL even has #version), but they miss #declare, and
			// light_source/camera are not keywords anywhere else -- it's fair
			// to assume though that any POVRAY scene must have a camera and
			// lightsource.
			var result float32

			if strings.Contains(text, "#version") {
				result += 0.05
			}

			if strings.Contains(text, "#declare") {
				result += 0.05
			}

			if strings.Contains(text, "camera") {
				result += 0.05
			}

			if strings.Contains(text, "light_source") {
				result += 0.1
			}

			return result
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (POVRay) Name() string {
	return heartbeat.LanguagePOVRay.StringChroma()
}
