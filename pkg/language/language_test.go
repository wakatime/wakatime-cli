package language_test

import (
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
	lang, err := language.Detect("testdata/codefiles/h_with_m_file/objective-c.m")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageObjectiveC, lang)
}

func TestDetect_ObjectiveC_M_FileInFolder(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/h_with_m_file/objective-c.h")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageObjectiveC, lang)
}

func TestDetect_ObjectiveCPP_MatchingHeader(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/h_with_mm_file/objective-cpp.mm")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageObjectiveCPP, lang)
}

func TestDetect_ObjectiveCPP_MM_FileInFolder(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/h_with_mm_file/objective-cpp.h")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageObjectiveCPP, lang)
}
