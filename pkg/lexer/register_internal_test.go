package lexer

import (
	"testing"

	"github.com/alecthomas/chroma/v2"
	chromalexers "github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterAll(t *testing.T) {
	require.NoError(t, RegisterAll())
	require.NotNil(t, chromalexers.Get(ADL{}.Name()))
	require.NotNil(t, chromalexers.Get(Zephir{}.Name()))
}

func TestCustomLexersTokenise(t *testing.T) {
	for _, customLexer := range customLexers() {
		t.Run(customLexer.Name(), func(t *testing.T) {
			lexer := customLexer.Lexer()
			require.NotNil(t, lexer)
			require.NotEmpty(t, customLexer.Name())

			iterator, err := lexer.Tokenise(nil, "sample\n")
			require.NoError(t, err)

			var tokens int
			for token := iterator(); token != chroma.EOF; token = iterator() {
				tokens++
			}

			assert.Positive(t, tokens)
		})
	}
}

func TestCustomLexerAnalysers(t *testing.T) {
	tests := map[string]struct {
		Lexer    Lexer
		Matching string
		Other    string
	}{
		"easytrieve": {
			Lexer: Easytrieve{},
			Matching: "* header\n" +
				"FILE INPUT\n" +
				"JOB INPUT INPUT\n" +
				"REPORT REPORT1\n" +
				"PROC REPORT\n" +
				"END-PROC\n",
			Other: "plain text",
		},
		"xml": {
			Lexer:    XML{},
			Matching: "<?xml version=\"1.0\"?><root></root>",
			Other:    "plain text",
		},
		"xslt": {
			Lexer:    XSLT{},
			Matching: "<?xml version=\"1.0\"?><xsl:stylesheet></xsl:stylesheet>",
			Other:    "<?xml version=\"1.0\"?><root></root>",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			lexer := test.Lexer.Lexer()
			require.NotNil(t, lexer)

			assert.Positive(t, lexer.AnalyseText(test.Matching))
			assert.Zero(t, lexer.AnalyseText(test.Other))
		})
	}
}

func TestCustomLexerAnalysersPlainText(t *testing.T) {
	for _, customLexer := range customLexers() {
		t.Run(customLexer.Name(), func(t *testing.T) {
			lexer := customLexer.Lexer()
			require.NotNil(t, lexer)

			assert.NotPanics(t, func() {
				_ = lexer.AnalyseText("plain text")
			})
		})
	}
}

func TestCustomLexerAnalysersMatchingSamples(t *testing.T) {
	tests := map[string]struct {
		Lexer Lexer
		Text  string
	}{
		"actionscript3": {
			Lexer: ActionScript3{},
			Text:  "var count:int = 1;",
		},
		"brainfuck": {
			Lexer: Brainfuck{},
			Text:  "++[>++<-].",
		},
		"coq": {
			Lexer: Coq{},
			Text:  "Theorem sample : True.",
		},
		"forth": {
			Lexer: Forth{},
			Text:  ": square dup * ;",
		},
		"fsharp": {
			Lexer: FSharp{},
			Text:  "let value = 1",
		},
		"gas": {
			Lexer: Gas{},
			Text:  ".section .text\n.globl _start",
		},
		"groff": {
			Lexer: Groff{},
			Text:  ".TH TEST 1",
		},
		"html": {
			Lexer: HTML{},
			Text:  "<!DOCTYPE html><html></html>",
		},
		"hy": {
			Lexer: Hy{},
			Text:  "(defn sample [] 1)",
		},
		"ini": {
			Lexer: INI{},
			Text:  "[settings]\nkey=value",
		},
		"make": {
			Lexer: Makefile{},
			Text:  "target:\n\tgo test ./...",
		},
		"matlab": {
			Lexer: Matlab{},
			Text:  "function y = f(x)\ny = x + 1;\nend",
		},
		"nasm": {
			Lexer: NASM{},
			Text:  "section .text\nglobal _start",
		},
		"objectivec": {
			Lexer: ObjectiveC{},
			Text:  "#import <Foundation/Foundation.h>\n@implementation Sample\n@end",
		},
		"prolog": {
			Lexer: Prolog{},
			Text:  "parent(alice, bob).",
		},
		"python": {
			Lexer: Python{},
			Text:  "import os\nprint(os.name)",
		},
		"python2": {
			Lexer: Python2{},
			Text:  "print 'hello'",
		},
		"qbasic": {
			Lexer: QBasic{},
			Text:  "10 PRINT \"HELLO\"",
		},
		"r": {
			Lexer: R{},
			Text:  "library(stats)\nvalue <- 1",
		},
		"tasm": {
			Lexer: TASM{},
			Text:  ".model small\n.code",
		},
		"turtle": {
			Lexer: Turtle{},
			Text:  "@prefix ex: <http://example.com/> .",
		},
		"vbnet": {
			Lexer: VBNet{},
			Text:  "Imports System\nModule Program\nEnd Module",
		},
		"verilog": {
			Lexer: Verilog{},
			Text:  "module top; endmodule",
		},
		"xml": {
			Lexer: XML{},
			Text:  "<?xml version=\"1.0\"?><root/>",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			lexer := test.Lexer.Lexer()
			require.NotNil(t, lexer)

			assert.GreaterOrEqual(t, lexer.AnalyseText(test.Text), float32(0))
		})
	}
}
