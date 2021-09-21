//go:build !windows

package pwd_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/pwd"

	"github.com/stretchr/testify/assert"
)

func TestGetpwuid(t *testing.T) {
	result := pwd.Getpwuid(0)
	assert.NotNil(t, result)

	assert.Equal(t, result.Name, "root")

	assert.Equal(t, result.UID, uint32(0))

	result = pwd.Getpwuid(1234556789) // uid which does not exit probably
	assert.Nil(t, result)
}
