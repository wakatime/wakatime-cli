package language

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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

func TestDetectChromaCustomizedFallbackBranches(t *testing.T) {
	tmpDir := t.TempDir()

	lang, weight, ok := detectChromaCustomized(t.Context(), filepath.Join(tmpDir, "unknown.nope"), false)
	assert.False(t, ok)
	assert.Equal(t, heartbeat.LanguageUnknown, lang)
	assert.Zero(t, weight)

	lang, weight, ok = detectChromaCustomized(t.Context(), filepath.Join(tmpDir, "missing.nope"), true)
	assert.False(t, ok)
	assert.Equal(t, heartbeat.LanguageUnknown, lang)
	assert.Zero(t, weight)

	emptyFile := filepath.Join(tmpDir, "empty.nope")
	require.NoError(t, os.WriteFile(emptyFile, nil, 0600))
	lang, weight, ok = detectChromaCustomized(t.Context(), emptyFile, true)
	assert.False(t, ok)
	assert.Equal(t, heartbeat.LanguageUnknown, lang)
	assert.Zero(t, weight)

	contentFile := filepath.Join(tmpDir, "script.nope")
	require.NoError(t, os.WriteFile(contentFile, []byte("#!/usr/bin/env python\nprint('x')\n"), 0600))

	lang, _, ok = detectChromaCustomized(t.Context(), contentFile, true)
	if ok {
		assert.NotEqual(t, heartbeat.LanguageUnknown, lang)
	}
}
