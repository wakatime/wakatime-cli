package regex

import (
	"testing"

	"github.com/dlclark/regexp2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegexp2Wrap_MatchString(t *testing.T) {
	tests := map[string]bool{
		"gopher":             false,
		"gophergopher":       true,
		"gophergophergopher": true,
	}

	for str, expected := range tests {
		t.Run(str, func(t *testing.T) {
			r2, err := regexp2.Compile(`(gopher){2}`, 0)
			require.NoError(t, err)

			r := &regexp2Wrap{
				rgx: r2,
			}

			assert.Equal(t, expected, r.MatchString(str))
		})
	}
}

func TestRegexp2Wrap_FindStringSubmatch(t *testing.T) {
	tests := map[string]struct {
		String   string
		Expected []string
	}{
		"submatch": {
			String:   "-axxxbyc-",
			Expected: []string{"axxxbyc", "xxx", "y"},
		},
		"empty submatch": {
			String:   "-abzc-",
			Expected: []string{"abzc", "", "z"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r2, err := regexp2.Compile(`a(x*)b(y|z)c`, 0)
			require.NoError(t, err)

			r := &regexp2Wrap{
				rgx: r2,
			}

			matches := r.FindStringSubmatch(test.String)

			assert.Equal(t, test.Expected, matches)
		})
	}
}
