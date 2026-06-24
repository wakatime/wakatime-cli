package remote

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldDownloadFileFallback(t *testing.T) {
	assert.True(t, shouldDownloadFileFallback(errors.New("connection reset")))
	assert.False(t, shouldDownloadFileFallback(errors.New("ssh: handshake failed: ssh: host key mismatch")))
}

func TestContains(t *testing.T) {
	assert.True(t, contains([]string{"GitHub.com", "example.com"}, "github.COM"))
	assert.False(t, contains([]string{"", "example.com"}, "github.com"))
}

func TestClientUser(t *testing.T) {
	assert.Equal(t, "explicit", Client{User: "explicit"}.user())
	assert.Empty(t, Client{}.user())
}
