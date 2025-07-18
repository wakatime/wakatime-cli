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
	v.Set("second", false)
	v.Set("third", true)

	value := vipertools.FirstNonEmptyBool(v, "first", "second", "third")
	assert.False(t, value)
}

func TestFirstNonEmptyBool_NonBool(t *testing.T) {
	v := viper.New()
	v.Set("first", "stringvalue")

	value := vipertools.FirstNonEmptyBool(v, "first")
	assert.False(t, value)
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
	value, ok := vipertools.FirstNonEmptyInt(v)
	require.False(t, ok)

	assert.Zero(t, value)
}

func TestFirstNonEmptyInt_NotFound(t *testing.T) {
	value, ok := vipertools.FirstNonEmptyInt(viper.New(), "key")
	require.False(t, ok)

	assert.Zero(t, value)
}

func TestFirstNonEmptyInt_EmptyInt(t *testing.T) {
	v := viper.New()
	v.Set("first", 0)

	value, ok := vipertools.FirstNonEmptyInt(v, "first")
	assert.True(t, ok)

	assert.Zero(t, value)
}

func TestFirstNonEmptyInt_StringValue(t *testing.T) {
	v := viper.New()
	v.Set("first", "stringvalue")

	value, ok := vipertools.FirstNonEmptyInt(v, "first")
	require.False(t, ok)

	assert.Zero(t, value)
}

func TestFirstNonEmptyString(t *testing.T) {
	v := viper.New()
	v.Set("second", "secret")
	v.Set("third", "ignored")

	value := vipertools.FirstNonEmptyString(v, "first", "second", "third")
	assert.Equal(t, "secret", value)
}

func TestFirstNonEmptyString_Empty(t *testing.T) {
	v := viper.New()
	v.Set("second", "")
	v.Set("third", "secret")

	value := vipertools.FirstNonEmptyString(v, "first", "second", "third")
	assert.Empty(t, value)
}

func TestFirstNonEmptyString_NilPointer(t *testing.T) {
	value := vipertools.FirstNonEmptyString(nil, "first")
	assert.Empty(t, value)
}

func TestFirstNonEmptyString_EmptyKeys(t *testing.T) {
	v := viper.New()
	value := vipertools.FirstNonEmptyString(v)
	assert.Empty(t, value)
}

func TestFirstNonEmptyString_NotFound(t *testing.T) {
	value := vipertools.FirstNonEmptyString(viper.New(), "key")
	assert.Empty(t, value)
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

func TestGetStringMapString(t *testing.T) {
	v := viper.New()
	v.Set("settings.github.com/wakatime", "value")
	v.Set("settings_foo.debug", "true")

	expected := map[string]string{
		"github.com/wakatime": "value",
	}

	m := vipertools.GetStringMapString(v, "settings")
	assert.Equal(t, expected, m)
}

func TestGetStringMapString_NotFound(t *testing.T) {
	v := viper.New()
	v.Set("settings.key", "value")

	m := vipertools.GetStringMapString(v, "internal")
	assert.Equal(t, map[string]string{}, m)
}
