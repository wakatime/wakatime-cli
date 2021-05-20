package vipertools_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirstNonEmptyBool(t *testing.T) {
	v := viper.New()
	v.Set("second", true)
	v.Set("third", false)

	value := vipertools.FirstNonEmptyBool(v, "first", "second", "third")
	assert.True(t, value)
}

func TestFirstNonEmptyBool_NilPointer(t *testing.T) {
	value := vipertools.FirstNonEmptyBool(nil, "first")
	assert.False(t, value)
}

func TestFirstNonEmptyBool_EmptyKeys(t *testing.T) {
	v := viper.New()
	value := vipertools.FirstNonEmptyBool(v)
	assert.False(t, value)
}

func TestFirstNonEmptyBool_NotFound(t *testing.T) {
	value := vipertools.FirstNonEmptyBool(viper.New(), "key")
	assert.False(t, value)
}

func TestFirstNonEmptyInt(t *testing.T) {
	v := viper.New()
	v.Set("second", 42)
	v.Set("third", 99)

	value, ok := vipertools.FirstNonEmptyInt(v, "first", "second", "third")
	require.True(t, ok)

	assert.Equal(t, 42, value)
}

func TestFirstNonEmptyInt_NilPointer(t *testing.T) {
	_, ok := vipertools.FirstNonEmptyInt(nil, "first")
	assert.False(t, ok)
}

func TestFirstNonEmptyInt_EmptyKeys(t *testing.T) {
	v := viper.New()
	_, ok := vipertools.FirstNonEmptyInt(v)
	assert.False(t, ok)
}

func TestFirstNonEmptyInt_NotFound(t *testing.T) {
	_, ok := vipertools.FirstNonEmptyInt(viper.New(), "key")
	assert.False(t, ok)
}

func TestFirstNonEmptyInt_EmptyInt(t *testing.T) {
	v := viper.New()
	v.Set("first", 0)
	_, ok := vipertools.FirstNonEmptyInt(v, "first")
	assert.False(t, ok)
}

func TestFirstNonEmptyInt_StringValue(t *testing.T) {
	v := viper.New()
	v.Set("first", "stringvalue")
	_, ok := vipertools.FirstNonEmptyInt(v, "first")
	assert.False(t, ok)
}

func TestFirstNonEmptyString(t *testing.T) {
	v := viper.New()
	v.Set("second", "secret")
	v.Set("third", "ignored")

	value, ok := vipertools.FirstNonEmptyString(v, "first", "second", "third")
	require.True(t, ok)
	assert.Equal(t, "secret", value)
}

func TestFirstNonEmptyString_NilPointer(t *testing.T) {
	_, ok := vipertools.FirstNonEmptyString(nil, "first")
	assert.False(t, ok)
}

func TestFirstNonEmptyString_EmptyKeys(t *testing.T) {
	v := viper.New()
	_, ok := vipertools.FirstNonEmptyString(v)
	assert.False(t, ok)
}

func TestFirstNonEmptyString_NotFound(t *testing.T) {
	_, ok := vipertools.FirstNonEmptyString(viper.New(), "key")
	assert.False(t, ok)
}

func TestGetString(t *testing.T) {
	v := viper.New()
	v.Set("some", "value")

	value := vipertools.GetString(v, "some")
	assert.Equal(t, "value", value)
}

func TestGetString_DoubleQuotes(t *testing.T) {
	v := viper.New()
	v.Set("some", "\"value\"")

	value := vipertools.GetString(v, "some")
	assert.Equal(t, "value", value)
}
