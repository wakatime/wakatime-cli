package output_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/output"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func outputTests() map[string]output.Output {
	return map[string]output.Output{
		"text":     output.TextOutput,
		"json":     output.JSONOutput,
		"raw-json": output.RawJSONOutput,
	}
}

func TestParseOutput(t *testing.T) {
	for value, out := range outputTests() {
		t.Run(value, func(t *testing.T) {
			parsed, err := output.Parse(value)
			require.NoError(t, err)

			assert.Equal(t, out, parsed)
		})
	}
}

func TestParseOutput_Invalid(t *testing.T) {
	_, err := output.Parse("invalid")
	require.Error(t, err)
}

func TestOutput_String(t *testing.T) {
	for value, out := range outputTests() {
		t.Run(value, func(t *testing.T) {
			assert.Equal(t, value, out.String())
		})
	}
}
