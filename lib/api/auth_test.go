package api_test

import (
	"testing"

	"github.com/alanhamlett/wakatime-cli/lib/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicAuth_HeaderValue(t *testing.T) {
	tests := map[string]struct {
		User, Secret, Expected string
	}{
		"standard": {
			User:     "john",
			Secret:   "secret",
			Expected: "Basic am9objpzZWNyZXQ=",
		},
		"apikey": {
			Secret:   "secret",
			Expected: "Basic c2VjcmV0",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			auth := api.BasicAuth{
				User:   test.User,
				Secret: test.Secret,
			}
			value, err := auth.HeaderValue()
			require.NoError(t, err)

			assert.Equal(t, test.Expected, value)
		})
	}
}

func TestBasicAuth_HeaderValue_MissingPassword(t *testing.T) {
	tests := map[string]struct {
		User, Secret string
	}{
		"only user": {
			User: "john",
		},
		"empty": {},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			auth := api.BasicAuth{
				User:   test.User,
				Secret: test.Secret,
			}
			_, err := auth.HeaderValue()
			require.Error(t, err)
		})
	}
}
