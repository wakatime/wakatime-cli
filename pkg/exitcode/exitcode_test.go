package exitcode_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/exitcode"

	"github.com/stretchr/testify/assert"
)

func TestErr_Error(t *testing.T) {
	var err error = exitcode.Err{Code: exitcode.ErrGeneric}
	assert.Equal(t, "1", err.Error())

	tests := map[string]struct {
		Code     int
		Expected string
	}{
		"success": {
			Code:     exitcode.Success,
			Expected: "0",
		},
		"api": {
			Code:     exitcode.ErrAPI,
			Expected: "102",
		},
		"auth": {
			Code:     exitcode.ErrAuth,
			Expected: "104",
		},
		"backoff": {
			Code:     exitcode.ErrBackoff,
			Expected: "112",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.Expected, exitcode.Err{Code: test.Code}.Error())
		})
	}
}
