package api_test

import (
	"errors"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestErrors(t *testing.T) {
	tests := map[string]struct {
		Err interface {
			error
			ExitCode() int
			Message() string
			SendDiagsOnErrors() bool
			ShouldLogError() bool
		}
		ExpectedError    string
		ExpectedExitCode int
		ExpectedMessage  string
		ExpectedLogError bool
	}{
		"api": {
			Err:              api.Err{Err: errors.New("server failed")},
			ExpectedError:    "server failed",
			ExpectedExitCode: exitcode.ErrAPI,
			ExpectedMessage:  "api error: server failed",
			ExpectedLogError: true,
		},
		"auth": {
			Err:              api.ErrAuth{Err: errors.New("unauthorized")},
			ExpectedError:    "unauthorized",
			ExpectedExitCode: exitcode.ErrAuth,
			ExpectedMessage:  "invalid api key... find yours at wakatime.com/api-key. unauthorized",
			ExpectedLogError: true,
		},
		"bad request": {
			Err:              api.ErrBadRequest{Err: errors.New("invalid payload")},
			ExpectedError:    "invalid payload",
			ExpectedExitCode: exitcode.ErrGeneric,
			ExpectedMessage:  "bad request: invalid payload",
			ExpectedLogError: true,
		},
		"backoff": {
			Err:              api.ErrBackoff{Err: errors.New("retry later")},
			ExpectedError:    "retry later",
			ExpectedExitCode: exitcode.ErrBackoff,
			ExpectedMessage:  "rate limited: retry later",
		},
		"timeout": {
			Err:              api.ErrTimeout{Err: errors.New("deadline exceeded")},
			ExpectedError:    "deadline exceeded",
			ExpectedExitCode: exitcode.ErrGeneric,
			ExpectedMessage:  "timeout: deadline exceeded",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.ExpectedError, test.Err.Error())
			assert.Equal(t, test.ExpectedExitCode, test.Err.ExitCode())
			assert.Equal(t, test.ExpectedMessage, test.Err.Message())
			assert.False(t, test.Err.SendDiagsOnErrors())
			assert.Equal(t, test.ExpectedLogError, test.Err.ShouldLogError())
		})
	}

	assert.Equal(t, int8(zapcore.DebugLevel), api.ErrBackoff{}.LogLevel())
	assert.Equal(t, int8(zapcore.DebugLevel), api.ErrTimeout{}.LogLevel())
}
