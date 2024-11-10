package log_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/stretchr/testify/assert"
)

func TestLog_IsMetricsEnabled(t *testing.T) {
	logger := log.New(nil, log.WithMetrics(true))

	assert.True(t, logger.IsMetricsEnabled())
}

func TestLog_IsMetricsEnabled_Disabled(t *testing.T) {
	logger := log.New(nil)

	assert.False(t, logger.IsMetricsEnabled())
}

func TestLog_IsVerboseEnabled(t *testing.T) {
	logger := log.New(nil, log.WithVerbose(true))

	assert.True(t, logger.IsVerboseEnabled())
}

func TestLog_IsVerboseEnabled_Disabled(t *testing.T) {
	logger := log.New(nil)

	assert.False(t, logger.IsVerboseEnabled())
}

func TestLog_SendDiagsOnErrors(t *testing.T) {
	logger := log.New(nil, log.WithSendDiagsOnErrors(true))

	assert.True(t, logger.SendDiagsOnErrors())
}

func TestLog_SendDiagsOnErrors_Disabled(t *testing.T) {
	logger := log.New(nil)

	assert.False(t, logger.SendDiagsOnErrors())
}
