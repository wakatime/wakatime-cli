package language_test

import (
	"fmt"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/language"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithDetection(t *testing.T) {
	tests := map[string]struct {
		Alternate string
		Override  string
		Expected  heartbeat.Language
	}{
		"alternate": {
			Alternate: "Go",
			Expected:  heartbeat.LanguageGo,
		},
		"override": {
			Alternate: "Go",
			Override:  "Python",
			Expected:  heartbeat.LanguagePython,
		},
		"empty": {
			Expected: heartbeat.LanguageUnknown,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			opt := language.WithDetection(language.Config{
				Alternate: test.Alternate,
				Override:  test.Override,
			})
			h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
				assert.Equal(t, []heartbeat.Heartbeat{
					{
						Language: test.Expected,
					},
				}, hh)

				return []heartbeat.Result{
					{
						Status: 201,
					},
				}, nil
			})

			result, err := h([]heartbeat.Heartbeat{{}})
			require.NoError(t, err)

			assert.Equal(t, []heartbeat.Result{
				{
					Status: 201,
				},
			}, result)
		})
	}
}

func TestDetect_HeaderFile_Corresponding_C_File(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/h_with_c_file/empty.h")
	require.NoError(t, err)
	assert.Equal(t, heartbeat.LanguageC, lang)
}

func TestDetect_HeaderFile_With_C_Files(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/h_with_any_c_file/empty.h")
	require.NoError(t, err)
	assert.Equal(t, heartbeat.LanguageC, lang)
}

func TestDetect_HeaderFile_With_C_And_CPP_Files(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/h_with_any_c_and_cpp_files/cpp.h")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageCPP, lang)
}

func TestDetect_HeaderFile_With_C_And_CXX_Files(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/h_with_any_c_and_cxx_files/cpp.h")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageCPP, lang)
}

func TestDetect_ObjectiveC_Over_Matlab_MatchingHeader(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/with_mat_file/objective-c.m")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageObjectiveC, lang)
}

func TestDetect_ObjectiveC_M_FileInFolder(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/with_mat_file/objective-c.h")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageObjectiveC, lang)
}

func TestDetect_ObjectiveCPP_MatchingHeader(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/with_mat_file/objective-cpp.mm")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageObjectiveCPP, lang)
}

func TestDetect_ObjectiveCPP_MM_FileInFolder(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/with_mat_file/objective-cpp.h")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageObjectiveCPP, lang)
}

func TestDetect_ObjectiveC(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/objective-c.m")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageObjectiveC, lang)
}

func TestDetect_Matlab_Over_ObjectiveC_Mat_FileInFolder(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/with_mat_file/empty.m")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageMatlab, lang)
}

func TestDetect_ObjectiveC_Over_Matlab_NonMatchingHeader(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/matlab_with_headers/empty.m")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageObjectiveC, lang)
}

func TestDetect_NonHeaderFile_C_FilesInFolder(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/py_with_c_files/see.py")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguagePython, lang)
}

func TestDetect_Perl_Over_Prolog(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/perl.pl")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguagePerl, lang)
}

func TestDetect_FSharp_Over_Forth(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/fsharp.fs")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageFSharp, lang)
}

func TestDetect_ChromaOverwrite(t *testing.T) {
	tests := map[string]heartbeat.Language{
		"testdata/codefiles/chroma_overwrite/cmakelists.txt": heartbeat.LanguageCMake,
		"testdata/codefiles/chroma_overwrite/go.mod":         heartbeat.LanguageGo,
	}

	for filepath, lang := range tests {
		t.Run(filepath, func(t *testing.T) {
			match, err := language.Detect(filepath)
			require.NoError(t, err)

			assert.Equal(t, lang, match, fmt.Sprintf("Got: %s, want: %s", match, lang))
		})
	}
}

func TestDetect_ChromaUnsupported(t *testing.T) {
	tests := map[string]heartbeat.Language{
		"testdata/codefiles/chroma_unsupported/empty.cfm":              heartbeat.LanguageColdfusionHTML,
		"testdata/codefiles/chroma_unsupported/empty.cfml":             heartbeat.LanguageColdfusionHTML,
		"testdata/codefiles/chroma_unsupported/crontab":                heartbeat.LanguageCrontab,
		"testdata/codefiles/chroma_unsupported/empty.pas":              heartbeat.LanguageDelphi,
		"testdata/codefiles/chroma_unsupported/empty.dpr":              heartbeat.LanguageDelphi,
		"testdata/codefiles/chroma_unsupported/empty.eex":              heartbeat.LanguageElixir,
		"testdata/codefiles/chroma_unsupported/empty.gs":               heartbeat.LanguageGosu,
		"testdata/codefiles/chroma_unsupported/empty.gsp":              heartbeat.LanguageGosu,
		"testdata/codefiles/chroma_unsupported/empty.gst":              heartbeat.LanguageGosu,
		"testdata/codefiles/chroma_unsupported/empty.gsx":              heartbeat.LanguageGosu,
		"testdata/codefiles/chroma_unsupported/empty.vark":             heartbeat.LanguageGosu,
		"testdata/codefiles/chroma_unsupported/empty.mjs":              heartbeat.LanguageJavaScript,
		"testdata/codefiles/chroma_unsupported/empty.jsx":              heartbeat.LanguageJSX,
		"testdata/codefiles/chroma_unsupported/empty.lasso":            heartbeat.LanguageLasso,
		"testdata/codefiles/chroma_unsupported/empty.lasso8":           heartbeat.LanguageLasso,
		"testdata/codefiles/chroma_unsupported/empty.lasso9":           heartbeat.LanguageLasso,
		"testdata/codefiles/chroma_unsupported/empty.less":             heartbeat.LanguageLess,
		"testdata/codefiles/chroma_unsupported/empty.liquid":           heartbeat.LanguageLiquid,
		"testdata/codefiles/chroma_unsupported/empty.marko":            heartbeat.LanguageMarko,
		"testdata/codefiles/chroma_unsupported/empty.mo":               heartbeat.LanguageModelica,
		"testdata/codefiles/chroma_unsupported/empty.mustache":         heartbeat.LanguageMustache,
		"testdata/codefiles/chroma_unsupported/empty.lsp":              heartbeat.LanguageNewLisp,
		"testdata/codefiles/chroma_unsupported/empty.kif":              heartbeat.LanguageNewLisp,
		"testdata/codefiles/chroma_unsupported/empty.nl":               heartbeat.LanguageNewLisp,
		"testdata/codefiles/chroma_unsupported/empty.pwn":              heartbeat.LanguagePawn,
		"testdata/codefiles/chroma_unsupported/empty.jade":             heartbeat.LanguagePug,
		"testdata/codefiles/chroma_unsupported/empty.pug":              heartbeat.LanguagePug,
		"testdata/codefiles/chroma_unsupported/empty.jy":               heartbeat.LanguagePython,
		"testdata/codefiles/chroma_unsupported/empty.bzl":              heartbeat.LanguagePython,
		"testdata/codefiles/chroma_unsupported/buck":                   heartbeat.LanguagePython,
		"testdata/codefiles/chroma_unsupported/build":                  heartbeat.LanguagePython,
		"testdata/codefiles/chroma_unsupported/build.bazel":            heartbeat.LanguagePython,
		"testdata/codefiles/chroma_unsupported/workspace":              heartbeat.LanguagePython,
		"testdata/codefiles/chroma_unsupported/empty.qml":              heartbeat.LanguageQML,
		"testdata/codefiles/chroma_unsupported/empty.qbs":              heartbeat.LanguageQML,
		"testdata/codefiles/chroma_unsupported/empty.spec":             heartbeat.LanguageRPMSpec,
		"testdata/codefiles/chroma_unsupported/empty.slim":             heartbeat.LanguageSlim,
		"testdata/codefiles/chroma_unsupported/empty.smali":            heartbeat.LanguageSmali,
		"testdata/codefiles/chroma_unsupported/empty.sketch":           heartbeat.LanguageSketchDrawing,
		"testdata/codefiles/chroma_unsupported/empty.svelte":           heartbeat.LanguageSvelte,
		"testdata/codefiles/chroma_unsupported/empty.sp":               heartbeat.LanguageSourcePawn,
		"testdata/codefiles/chroma_unsupported/empty.sublime-settings": heartbeat.LanguageSublimeTextConfig,
		"testdata/codefiles/chroma_unsupported/empty.swg":              heartbeat.LanguageSWIG,
		"testdata/codefiles/chroma_unsupported/empty.i":                heartbeat.LanguageSWIG,
		"testdata/codefiles/chroma_unsupported/pipfile":                heartbeat.LanguageTOML,
		"testdata/codefiles/chroma_unsupported/poetry.lock":            heartbeat.LanguageTOML,
		"testdata/codefiles/chroma_unsupported/empty.twig":             heartbeat.LanguageTwig,
		"testdata/codefiles/chroma_unsupported/empty.vcl":              heartbeat.LanguageVCL,
		"testdata/codefiles/chroma_unsupported/empty.vm":               heartbeat.LanguageVelocity,
		"testdata/codefiles/chroma_unsupported/empty.fhtml":            heartbeat.LanguageVelocity,
		"testdata/codefiles/chroma_unsupported/empty.vue":              heartbeat.LanguageVueJS,
		"testdata/codefiles/chroma_unsupported/empty.xaml":             heartbeat.LanguageXAML,
		"testdata/codefiles/chroma_unsupported/empty.xpl":              heartbeat.LanguageXSLT,
	}

	for filepath, lang := range tests {
		t.Run(filepath, func(t *testing.T) {
			match, err := language.Detect(filepath)
			require.NoError(t, err)

			assert.Equal(t, lang, match, fmt.Sprintf("Got: %s, want: %s", match, lang))
		})
	}
}
