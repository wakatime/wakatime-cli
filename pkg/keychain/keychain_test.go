package keychain_test

import (
	"errors"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/keychain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func TestWriteAPIKey(t *testing.T) {
	keyring.MockInit()

	err := keychain.WriteAPIKey("00000000-0000-4000-8000-000000000000")
	require.NoError(t, err)

	stored, err := keyring.Get(keychain.Service, keychain.User)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-000000000000", stored)
}

func TestReadAPIKey(t *testing.T) {
	keyring.MockInit()

	err := keyring.Set(keychain.Service, keychain.User, " 00000000-0000-4000-8000-000000000000\n")
	require.NoError(t, err)

	apiKey, err := keychain.ReadAPIKey()
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-000000000000", apiKey)
}

func TestReadAPIKey_NotFound(t *testing.T) {
	keyring.MockInit()

	_, err := keychain.ReadAPIKey()

	assert.ErrorIs(t, err, keychain.ErrNotFound)
}

func TestReadAPIKey_Err(t *testing.T) {
	keyring.MockInitWithError(errors.New("keychain locked"))

	_, err := keychain.ReadAPIKey()

	assert.EqualError(t, err, "keychain locked")
}
