package lexer

import (
	"testing"

	chromalexers "github.com/alecthomas/chroma/v2/lexers"
	"github.com/stretchr/testify/require"
)

func TestRegisterAll(t *testing.T) {
	require.NoError(t, RegisterAll())
	require.NotNil(t, chromalexers.Get(ADL{}.Name()))
	require.NotNil(t, chromalexers.Get(Zephir{}.Name()))
}
