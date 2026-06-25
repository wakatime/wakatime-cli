package deps

import (
	"context"
	"errors"
	"testing"

	"github.com/alecthomas/chroma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type panicParser struct{}

func (panicParser) Parse(context.Context, string) ([]string, error) {
	panic("lexer failed")
}

type dependenciesParser struct{}

func (dependenciesParser) Parse(context.Context, string) ([]string, error) {
	return []string{"requests", "requests"}, nil
}

type errorParser struct{}

func (errorParser) Parse(context.Context, string) ([]string, error) {
	return nil, errors.New("failed")
}

func TestParseDependencies_Panic(t *testing.T) {
	dependencies, err := parseDependencies(t.Context(), panicParser{}, "testdata/python.py")

	require.Error(t, err)
	assert.Nil(t, dependencies)
	assert.Contains(t, err.Error(), "panicked: lexer failed")
}

func TestParseDependencies_Success(t *testing.T) {
	dependencies, err := parseDependencies(t.Context(), dependenciesParser{}, "testdata/python.py")

	require.NoError(t, err)
	assert.Equal(t, []string{"requests", "requests"}, dependencies)
}

func TestParseDependencies_Error(t *testing.T) {
	dependencies, err := parseDependencies(t.Context(), errorParser{}, "testdata/python.py")

	require.EqualError(t, err, "failed")
	assert.Nil(t, dependencies)
}

func TestParserStateMachineBranches(t *testing.T) {
	t.Run("javascript", func(t *testing.T) {
		p := &ParserJavaScript{}
		p.init()

		p.processKeywordReserved("import")
		p.processPunctuation(";")
		assert.Equal(t, StateJavaScriptUnknown, p.State)

		p.processKeywordReserved("import")
		p.processLiteralStringSingle(`"../pkg/example.ts"`)
		assert.Equal(t, []string{"example"}, p.Output)

		p.processKeywordReserved("const")
		assert.Equal(t, StateJavaScriptUnknown, p.State)
	})

	t.Run("python", func(t *testing.T) {
		p := &ParserPython{}
		p.init()

		p.append("")
		p.append("os")
		p.append("requests.sessions")
		assert.Equal(t, []string{"requests"}, p.Output)

		p.processKeywordNamespace("import")
		p.processNameNamespace("numpy")
		p.processOperator(",")
		assert.Equal(t, []string{"requests", "numpy"}, p.Output)

		p.processKeywordNamespace("import")
		p.processNameNamespace("pandas")
		p.processKeyword("as")
		assert.Equal(t, []string{"requests", "numpy", "pandas"}, p.Output)

		p.processKeywordNamespace("from")
		p.processNameNamespace("collections")
		p.processText("\n")
		assert.Equal(t, StatePythonUnknown, p.State)

		p.processKeywordNamespace("class")
		p.processOperator(".")
		p.processText(" ")
		assert.Equal(t, StatePythonUnknown, p.State)
	})

	t.Run("go", func(t *testing.T) {
		p := &ParserGo{}
		p.init()

		p.processKeywordNamespace("import")
		p.processLiteralString(`"fmt"`)
		assert.Empty(t, p.Output)

		p.processLiteralString(`"github.com/wakatime/wakatime-cli"`)
		assert.Equal(t, []string{"github.com/wakatime/wakatime-cli"}, p.Output)

		p.processNameFunction()
		assert.Equal(t, StateGoUnknown, p.State)

		p.processKeywordNamespace("func")
		assert.Equal(t, StateGoUnknown, p.State)
	})

	t.Run("vbnet", func(t *testing.T) {
		p := &ParserVbNet{}
		p.init()

		p.append("")
		p.append("System.Text")
		assert.Empty(t, p.Output)

		p.processKeyword("Imports")
		p.processName("Newtonsoft")
		p.processName(".")
		p.processName("Json")
		p.processText("\n")
		assert.Equal(t, []string{"Newtonsoft"}, p.Output)

		p.processKeyword("Imports")
		p.processName("Alias")
		p.processOperator("=")
		p.processName("Custom")
		p.processPunctuation(".")
		p.processName("Lib")
		p.processText("\n")
		assert.Equal(t, []string{"Newtonsoft", "Custom"}, p.Output)

		p.processOperator("=")
		p.processPunctuation(".")
		assert.Equal(t, StateVbNetUnknown, p.State)
	})

	t.Run("php", func(t *testing.T) {
		p := &ParserPHP{}
		p.init()

		p.append("")
		p.append("app")
		assert.Empty(t, p.Output)

		p.processKeyword("include")
		p.processLiteralStringSingle("vendor/autoload.php")
		assert.Contains(t, p.Output, "vendor/autoload.php")

		p.processKeyword("include")
		p.processLiteralStringSingle(`'`)
		assert.Equal(t, StatePHPInclude, p.State)

		p.processLiteralStringDouble("config/bootstrap.php")
		assert.Contains(t, p.Output, "'config/bootstrap.php'")

		p.processKeyword("use")
		p.processNameOther(`Vendor\Package\ClassName`)
		assert.Contains(t, p.Output, "Vendor")

		p.processKeyword("use")
		p.processKeyword("function")
		p.processNameFunction(`Helpers\run`)
		assert.Contains(t, p.Output, "Helpers")

		p.processKeyword("as")
		p.processPunctuation(",")
		assert.Equal(t, StatePHPUse, p.State)

		p.processPunctuation(";")
		assert.Equal(t, StatePHPUnknown, p.State)

		p.processToken(chroma.Token{Type: chroma.Comment, Value: "x"})
		assert.Equal(t, StatePHPUnknown, p.State)
	})

	t.Run("c", func(t *testing.T) {
		p := &ParserC{}
		p.init()

		p.append("")
		p.append("stdio.h")
		assert.Empty(t, p.Output)

		p.processCommentPreprocFile("<ignored.h>")
		assert.Empty(t, p.Output)

		p.processCommentPreproc(" include")
		p.processCommentPreprocFile("<vendor/lib.h>")
		assert.Equal(t, []string{"vendor"}, p.Output)
		assert.Equal(t, StateCUnknown, p.State)

		p.processCommentPreproc("include")
		p.processCommentPreprocFile("#")
		assert.Equal(t, []string{"vendor"}, p.Output)
	})

	t.Run("cpp", func(t *testing.T) {
		p := &ParserCPP{}
		p.init()

		p.append("")
		p.append("iostream")
		assert.Empty(t, p.Output)

		p.processCommentPreprocFile("<ignored.hpp>")
		assert.Empty(t, p.Output)

		p.processCommentPreproc("include")
		p.processCommentPreprocFile(`"boost/asio.hpp"`)
		assert.Equal(t, []string{"boost"}, p.Output)
		assert.Equal(t, StateCPPUnknown, p.State)

		p.processCommentPreproc("include")
		p.processCommentPreprocFile("\n")
		assert.Equal(t, []string{"boost"}, p.Output)
	})

	t.Run("java", func(t *testing.T) {
		p := &ParserJava{}
		p.init()

		p.append("java.util")
		p.append(" ")
		assert.Empty(t, p.Output)

		p.processKeywordNamespace("class")
		assert.Equal(t, StateJavaUnknown, p.State)

		p.processKeywordNamespace("import")
		p.processNameNamespace("static")
		p.processNameNamespace("com")
		p.processPunctuation(".")
		p.processName("example")
		p.processPunctuation(".")
		p.processNameAttribute("*")
		p.processPunctuation(";")
		assert.Equal(t, []string{"example"}, p.Output)

		p.State = StateJavaImportFinished
		p.processKeywordNamespace("org.example.library")
		assert.Equal(t, []string{"example", "example.library"}, p.Output)
	})

	t.Run("json", func(t *testing.T) {
		p := &ParserJSON{}
		p.init()

		p.processFilename("package.json")
		assert.Equal(t, []string{"npm"}, p.Output)

		p.processPunctuation("{")
		p.processNameTag(`"dependencies"`)
		p.processPunctuation("{")
		p.processNameTag(`"github.com/wakatime/wakatime-cli"`)
		p.processPunctuation("}")
		assert.Equal(t, StateJSONUnknown, p.State)
		assert.Equal(t, []string{"npm", "github.com/wakatime/wakatime-cli"}, p.Output)
	})

	t.Run("csharp", func(t *testing.T) {
		p := &ParserCSharp{}
		p.init()

		p.append("")
		p.append("System.Text")
		assert.Empty(t, p.Output)

		p.processName("Ignored")
		assert.Equal(t, StateCSharpUnknown, p.State)

		p.processKeyword("using")
		p.processName("Alias")
		p.processPunctuation("=")
		p.processName("Newtonsoft")
		p.processPunctuation(".")
		p.processName("Json")
		p.processPunctuation(";")
		assert.Equal(t, []string{"Newtonsoft"}, p.Output)
		assert.Equal(t, StateCSharpUnknown, p.State)
	})

	t.Run("kotlin", func(t *testing.T) {
		p := &ParserKotlin{}
		p.init()

		p.append("java.util")
		p.append("kotlinx.coroutines.*")
		assert.Equal(t, []string{"kotlinx.coroutines"}, p.Output)

		p.processKeyword("import")
		p.processNameNamespace("com.example.lib")
		assert.Equal(t, []string{"kotlinx.coroutines", "com.example"}, p.Output)

		p.processKeyword("class")
		assert.Equal(t, StateKotlinUnknown, p.State)
		p.processToken(chroma.Token{Type: chroma.Punctuation, Value: ";"})
		assert.Equal(t, StateKotlinUnknown, p.State)
	})

	t.Run("rust", func(t *testing.T) {
		p := &ParserRust{}
		p.init()

		p.processKeyword("crate")
		p.processName("ignored")
		assert.Empty(t, p.Output)

		p.processKeyword("extern")
		p.processKeyword("crate")
		p.processName(" serde ")
		assert.Equal(t, []string{"serde"}, p.Output)
		assert.Equal(t, StateRustUnknown, p.State)
	})

	t.Run("swift", func(t *testing.T) {
		p := &ParserSwift{}
		p.init()

		p.append("Foundation")
		assert.Empty(t, p.Output)

		p.processKeywordDeclaration("import")
		p.processNameClass("Combine")
		assert.Equal(t, []string{"Combine"}, p.Output)

		p.processKeywordDeclaration("class")
		assert.Equal(t, StateSwiftUnknown, p.State)
		p.processNameClass("Ignored")
		assert.Equal(t, StateSwiftUnknown, p.State)
	})
}
