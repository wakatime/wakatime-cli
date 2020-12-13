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
		Plugin    string
		Expected  heartbeat.Language
	}{
		"alternate": {
			Alternate: "Go",
			Plugin:    "default",
			Expected:  heartbeat.LanguageGo,
		},
		"override": {
			Alternate: "Go",
			Override:  "Python",
			Plugin:    "default",
			Expected:  heartbeat.LanguagePython,
		},
		"alternate - default plugin specific language": {
			Alternate: "ld-script",
			Plugin:    "default",
			Expected:  heartbeat.LanguageLinkerScript,
		},
		"override - default plugin specific language": {
			Alternate: "Go",
			Override:  "ld-script",
			Plugin:    "default",
			Expected:  heartbeat.LanguageLinkerScript,
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

			userAgent := fmt.Sprintf("wakatime/0.0.1 (linux-4.13.0-38-generic-x86_64) go1.15.3 %s", test.Plugin)
			h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
				assert.Equal(t, []heartbeat.Heartbeat{
					{
						Language:  test.Expected,
						UserAgent: userAgent,
					},
				}, hh)

				return []heartbeat.Result{
					{
						Status: 201,
					},
				}, nil
			})

			result, err := h([]heartbeat.Heartbeat{{
				UserAgent: userAgent,
			}})
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

func TestDetect_GoMod(t *testing.T) {
	lang, err := language.Detect("/path/to/go.mod")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageGo, lang)
}
