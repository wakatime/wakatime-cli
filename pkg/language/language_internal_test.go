package language

import (
	"context"
	"errors"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func panicDetector(context.Context, string, bool) (heartbeat.Language, error) {
	panic("lexer failed")
}

func languageDetector(context.Context, string, bool) (heartbeat.Language, error) {
	return heartbeat.LanguagePython, nil
}

func errorDetector(context.Context, string, bool) (heartbeat.Language, error) {
	return heartbeat.LanguageUnknown, errors.New("failed")
}

func TestDetectLanguage_Panic(t *testing.T) {
	language, err := detectLanguage(t.Context(), panicDetector, "testdata/codefiles/python.py", false)

	require.Error(t, err)
	assert.Equal(t, heartbeat.LanguageUnknown, language)
	assert.Contains(t, err.Error(), "panicked: lexer failed")
}

func TestDetectLanguage_Success(t *testing.T) {
	language, err := detectLanguage(t.Context(), languageDetector, "testdata/codefiles/python.py", false)

	require.NoError(t, err)
	assert.Equal(t, heartbeat.LanguagePython, language)
}

func TestDetectLanguage_Error(t *testing.T) {
	language, err := detectLanguage(t.Context(), errorDetector, "testdata/codefiles/python.py", false)

	require.EqualError(t, err, "failed")
	assert.Equal(t, heartbeat.LanguageUnknown, language)
}
