package keychain

import (
	"errors"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	// Service is the service name the api key is stored under in the OS keychain.
	Service = "wakatime"
	// User is the account name the api key is stored under in the OS keychain.
	User = "api_key"
	// StorageKeychain is the value of settings.api_key_storage that enables reading the api key from the OS keychain.
	StorageKeychain = "keychain"
)

// ErrNotFound is returned when no api key is stored in the OS keychain.
var ErrNotFound = errors.New("api key not found in keychain")

// ReadAPIKey reads the api key from the OS keychain. It uses the macOS Keychain,
// the Windows Credential Manager or the Secret Service on Linux.
func ReadAPIKey() (string, error) {
	apiKey, err := keyring.Get(Service, User)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrNotFound
		}

		return "", err
	}

	return strings.TrimSpace(apiKey), nil
}

// WriteAPIKey stores the api key in the OS keychain, replacing any existing value.
func WriteAPIKey(apiKey string) error {
	return keyring.Set(Service, User, apiKey)
}
