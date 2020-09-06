package language_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/language"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetect_ByFileExtension(t *testing.T) {
	tests := map[string]struct {
		Filepaths []string
		Expected  string
	}{
		"asp": {
			Filepaths: []string{"testdata/codefiles/file.asp"},
			Expected:  "ASP",
		},
		"assembly not gas": {
			Filepaths: []string{"testdata/codefiles/gas.s"},
			Expected:  "Assembly",
		},
		"bash": {
			Filepaths: []string{
				"testdata/codefiles/file.bash",
				"testdata/codefiles/file.ksh",
				"testdata/codefiles/file.sh",
				"testdata/codefiles/file.zsh",
			},
			Expected: "Bash",
		},
		"c": {
			Filepaths: []string{"testdata/codefiles/c_only/foo.c"},
			Expected:  "C",
		},
		"c++": {
			Filepaths: []string{"testdata/codefiles/c_and_cpp/foo.cpp"},
			Expected:  "C++",
		},
		"c++ 2": {
			Filepaths: []string{"testdata/codefiles/c_and_cxx/foo.cxx"},
			Expected:  "C++",
		},
		"c sharp": {
			Filepaths: []string{"testdata/codefiles/csharp/seesharp.cs"},
			Expected:  "C#",
		},
		"coldfusion": {
			Filepaths: []string{"testdata/codefiles/coldfusion.cfm"},
			Expected:  "ColdFusion",
		},
		"cshtml": {
			Filepaths: []string{"testdata/codefiles/file.cshtml"},
			Expected:  "CSHTML",
		},
		"delphi": {
			Filepaths: []string{"testdata/codefiles/file.pas"},
			Expected:  "Delphi",
		},
		"elm": {
			Filepaths: []string{"testdata/codefiles/elm.elm"},
			Expected:  "Elm",
		},
		"f sharp not forth": {
			Filepaths: []string{"testdata/codefiles/fsharp.fs"},
			Expected:  "F#",
		},
		"golang": {
			Filepaths: []string{"testdata/codefiles/go.go"},
			Expected:  "Go",
		},
		"golang modfile": {
			Filepaths: []string{"testdata/codefiles/go.mod"},
			Expected:  "Go",
		},
		"gosu": {
			Filepaths: []string{
				"testdata/codefiles/file.gs",
				"testdata/codefiles/file.gsp",
				"testdata/codefiles/file.gst",
				"testdata/codefiles/file.gsx",
			},
			Expected: "Gosu",
		},
		"haml": {
			Filepaths: []string{"testdata/codefiles/file.haml"},
			Expected:  "Haml",
		},
		"haskell": {
			Filepaths: []string{"testdata/codefiles/haskell.hs"},
			Expected:  "Haskell",
		},
		"haxe": {
			Filepaths: []string{"testdata/codefiles/haxe.hx"},
			Expected:  "Haxe",
		},
		"html": {
			Filepaths: []string{"testdata/codefiles/html.html"},
			Expected:  "HTML",
		},
		"jade": {
			Filepaths: []string{"testdata/codefiles/file.jade"},
			Expected:  "Jade",
		},
		"java": {
			Filepaths: []string{"testdata/codefiles/java.java"},
			Expected:  "Java",
		},
		"javascript": {
			Filepaths: []string{"testdata/codefiles/es6.js"},
			Expected:  "JavaScript",
		},
		"javascript module extension": {
			Filepaths: []string{"testdata/codefiles/javascript_module.mjs"},
			Expected:  "JavaScript",
		},
		"json": {
			Filepaths: []string{"testdata/codefiles/bower.json"},
			Expected:  "JSON",
		},
		"jsx": {
			Filepaths: []string{"testdata/codefiles/file.jsx"},
			Expected:  "JSX",
		},
		"kotlin": {
			Filepaths: []string{"testdata/codefiles/kotlin.kt"},
			Expected:  "Kotlin",
		},
		"less": {
			Filepaths: []string{"testdata/codefiles/file.less"},
			Expected:  "LESS",
		},
		"markdown": {
			Filepaths: []string{"testdata/codefiles/file.md"},
			Expected:  "Markdown",
		},
		"mustache": {
			Filepaths: []string{"testdata/codefiles/file.mustache"},
			Expected:  "Mustache",
		},
		"perl not prolog": {
			Filepaths: []string{"testdata/codefiles/perl.pl"},
			Expected:  "Perl",
		},
		"php": {
			Filepaths: []string{"testdata/codefiles/php.php"},
			Expected:  "PHP",
		},
		"python": {
			Filepaths: []string{"testdata/codefiles/python.py"},
			Expected:  "Python",
		},
		"restructuredtext": {
			Filepaths: []string{"testdata/codefiles/file.rst"},
			Expected:  "reStructuredText",
		},
		"rust": {
			Filepaths: []string{"testdata/codefiles/rust.rs"},
			Expected:  "Rust",
		},
		"scala": {
			Filepaths: []string{"testdata/codefiles/scala.scala"},
			Expected:  "Scala",
		},
		"swift": {
			Filepaths: []string{"testdata/codefiles/swift.swift"},
			Expected:  "Swift",
		},
		"textfile": {
			Filepaths: []string{"testdata/codefiles/emptyfile.txt"},
			Expected:  "plaintext",
		},
		"typescript": {
			Filepaths: []string{"testdata/codefiles/typescript.ts"},
			Expected:  "TypeScript",
		},
		"typoscript": {
			Filepaths: []string{"testdata/codefiles/typoscript.typoscript"},
			Expected:  "TypoScript",
		},
		"xaml": {
			Filepaths: []string{"testdata/codefiles/file.xaml"},
			Expected:  "XAML",
		},
		"case-insensitive": {
			Filepaths: []string{"testdata/codefiles/FILE.GO"},
			Expected:  "Go",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for _, filepath := range test.Filepaths {
				lang, err := language.Detect(filepath)
				require.NoError(t, err)

				assert.Equal(t, test.Expected, lang)
			}
		})
	}
}
