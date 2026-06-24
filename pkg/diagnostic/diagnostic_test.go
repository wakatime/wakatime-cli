package diagnostic_test

import (
	"errors"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/diagnostic"

	"github.com/stretchr/testify/assert"
)

func TestConstructors(t *testing.T) {
	tests := map[string]struct {
		Diagnostic diagnostic.Diagnostic
		Expected   diagnostic.Diagnostic
	}{
		"error": {
			Diagnostic: diagnostic.Error(errors.New("boom")),
			Expected: diagnostic.Diagnostic{
				Type:  diagnostic.TypeError,
				Value: "boom",
			},
		},
		"logs": {
			Diagnostic: diagnostic.Logs("debug output"),
			Expected: diagnostic.Diagnostic{
				Type:  diagnostic.TypeLogs,
				Value: "debug output",
			},
		},
		"stack": {
			Diagnostic: diagnostic.Stack("goroutine 1"),
			Expected: diagnostic.Diagnostic{
				Type:  diagnostic.TypeStack,
				Value: "goroutine 1",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.Expected, test.Diagnostic)
		})
	}
}
