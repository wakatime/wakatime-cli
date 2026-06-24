package offline_test

import (
	"errors"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/offline"

	"github.com/stretchr/testify/assert"
)

func TestErrOpenDB(t *testing.T) {
	err := offline.ErrOpenDB{Err: errors.New("permission denied")}

	assert.Equal(t, "permission denied", err.Error())
	assert.Equal(t, "failed to open db file: permission denied", err.Message())
	assert.Equal(t, exitcode.Success, err.ExitCode())
	assert.True(t, err.SendDiagsOnErrors())
	assert.True(t, err.ShouldLogError())
}
