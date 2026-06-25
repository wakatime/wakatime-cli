package log_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
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

func TestLoggerOutputAndSetOutput(t *testing.T) {
	var first bytes.Buffer

	logger := log.New(&first)

	assert.Equal(t, &first, logger.Output())

	logger.Infoln("first")
	assert.Contains(t, first.String(), "first")

	var second bytes.Buffer
	logger.SetOutput(&second)

	assert.Equal(t, &second, logger.Output())

	logger.Infoln("second")
	assert.Contains(t, second.String(), "second")
	assert.NotContains(t, first.String(), "second")
}

func TestLoggerLoggingMethods(t *testing.T) {
	var buffer bytes.Buffer

	logger := log.New(&buffer, log.WithVerbose(true))

	logger.Log(zapcore.InfoLevel, "plain log")
	logger.Logf(zapcore.InfoLevel, "formatted %s", "log")
	logger.Debugf("debug %s", "format")
	logger.Infof("info %s", "format")
	logger.Warnf("warn %s", "format")
	logger.Errorf("error %s", "format")
	logger.Debugln("debug line")
	logger.Infoln("info line")
	logger.Warnln("warn line")
	logger.Errorln("error line")
	logger.WithField("component", "test")
	logger.Infoln("fielded")

	output := buffer.String()
	assert.Contains(t, output, "plain log")
	assert.Contains(t, output, "formatted log")
	assert.Contains(t, output, "debug format")
	assert.Contains(t, output, "info format")
	assert.Contains(t, output, "warn format")
	assert.Contains(t, output, "error format")
	assert.Contains(t, output, "debug line")
	assert.Contains(t, output, "info line")
	assert.Contains(t, output, "warn line")
	assert.Contains(t, output, "error line")
	assert.Contains(t, output, "fielded")
	assert.Contains(t, output, `"component":"test"`)
}

func TestLoggerContextHelpers(t *testing.T) {
	var buffer bytes.Buffer

	logger := log.New(&buffer)

	ctx := log.ToContext(context.Background(), logger)

	assert.Equal(t, logger, log.Extract(ctx))

	log.AddField(ctx, "request_id", "abc")
	logger.Infoln("message")

	assert.Contains(t, buffer.String(), `"request_id":"abc"`)
	assert.NotNil(t, log.Extract(context.Background()))
}

func TestLoggerFlushClosesOutput(t *testing.T) {
	writer := &closeTrackingWriter{}
	logger := log.New(writer)

	logger.Flush()

	assert.True(t, writer.closed)
}

func TestDynamicWriteSyncer(t *testing.T) {
	var first bytes.Buffer

	syncer := log.NewDynamicWriteSyncer(zapcore.AddSync(&first))

	n, err := syncer.Write([]byte("first"))
	require.NoError(t, err)
	assert.Equal(t, len("first"), n)
	assert.Equal(t, "first", first.String())

	var second bytes.Buffer
	syncer.SetWriter(zapcore.AddSync(&second))

	n, err = syncer.Write([]byte("second"))
	require.NoError(t, err)
	assert.Equal(t, len("second"), n)
	assert.Equal(t, "second", second.String())
	assert.Equal(t, "first", first.String())

	require.NoError(t, syncer.Sync())
}

type closeTrackingWriter struct {
	closed bool
}

func (*closeTrackingWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (w *closeTrackingWriter) Close() error {
	w.closed = true
	return nil
}
