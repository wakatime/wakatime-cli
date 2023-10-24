package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var (
	swigAnalyserDirectivesRe = regexp.MustCompile(`(?m)^\s*(%[a-z_][a-z0-9_]*)`)
	// nolint:gochecknoglobals
	swigAnalyserDirectives = map[string]struct{}{
		// Most common directives
		`%apply`:      {},
		`%define`:     {},
		`%director`:   {},
		`%enddef`:     {},
		`%exception`:  {},
		`%extend`:     {},
		`%feature`:    {},
		`%fragment`:   {},
		`%ignore`:     {},
		`%immutable`:  {},
		`%import`:     {},
		`%include`:    {},
		`%inline`:     {},
		`%insert`:     {},
		`%module`:     {},
		`%newobject`:  {},
		`%nspace`:     {},
		`%pragma`:     {},
		`%rename`:     {},
		`%shared_ptr`: {},
		`%template`:   {},
		`%typecheck`:  {},
		`%typemap`:    {},
		// Less common directives
		`%arg`:                  {},
		`%attribute`:            {},
		`%bang`:                 {},
		`%begin`:                {},
		`%callback`:             {},
		`%catches`:              {},
		`%clear`:                {},
		`%constant`:             {},
		`%copyctor`:             {},
		`%csconst`:              {},
		`%csconstvalue`:         {},
		`%csenum`:               {},
		`%csmethodmodifiers`:    {},
		`%csnothrowexception`:   {},
		`%default`:              {},
		`%defaultctor`:          {},
		`%defaultdtor`:          {},
		`%defined`:              {},
		`%delete`:               {},
		`%delobject`:            {},
		`%descriptor`:           {},
		`%exceptionclass`:       {},
		`%exceptionvar`:         {},
		`%extend_smart_pointer`: {},
		`%fragments`:            {},
		`%header`:               {},
		`%ifcplusplus`:          {},
		`%ignorewarn`:           {},
		`%implicit`:             {},
		`%implicitconv`:         {},
		`%init`:                 {},
		`%javaconst`:            {},
		`%javaconstvalue`:       {},
		`%javaenum`:             {},
		`%javaexception`:        {},
		`%javamethodmodifiers`:  {},
		`%kwargs`:               {},
		`%luacode`:              {},
		`%mutable`:              {},
		`%naturalvar`:           {},
		`%nestedworkaround`:     {},
		`%perlcode`:             {},
		`%pythonabc`:            {},
		`%pythonappend`:         {},
		`%pythoncallback`:       {},
		`%pythoncode`:           {},
		`%pythondynamic`:        {},
		`%pythonmaybecall`:      {},
		`%pythonnondynamic`:     {},
		`%pythonprepend`:        {},
		`%refobject`:            {},
		`%shadow`:               {},
		`%sizeof`:               {},
		`%trackobjects`:         {},
		`%types`:                {},
		`%unrefobject`:          {},
		`%varargs`:              {},
		`%warn`:                 {},
		`%warnfilter`:           {},
	}
)

// SWIG lexer.
type SWIG struct{}

// Lexer returns the lexer.
func (l SWIG) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"swig"},
			Filenames: []string{"*.swg", "*.i"},
			MimeTypes: []string{"text/swig"},
			// Lower than C/C++ and Objective C/C++
			Priority: 0.04,
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		var result float32

		// Search for SWIG directives, which are conventionally at the beginning of
		// a line. The probability of them being within a line is low, so let another
		// lexer win in this case.
		matches := swigAnalyserDirectivesRe.FindAllString(text, -1)

		for _, m := range matches {
			if _, ok := swigAnalyserDirectives[m]; ok {
				result = 0.98
				break
			}

			// Fraction higher than MatlabLexer
			result = 0.91
		}

		return result
	})

	return lexer
}

// Name returns the name of the lexer.
func (SWIG) Name() string {
	return heartbeat.LanguageSWIG.StringChroma()
}
