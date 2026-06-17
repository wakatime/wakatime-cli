package ai

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type panicParser struct{}

func (panicParser) Parse(context.Context) (Heartbeats, error) {
	panic("OOM")
}

func (panicParser) Name() string {
	return "panic"
}

type heartbeatsParser struct{}

func (heartbeatsParser) Parse(context.Context) (Heartbeats, error) {
	return make(Heartbeats, 2), nil
}

func (heartbeatsParser) Name() string {
	return "heartbeats"
}

type errorParser struct{}

func (errorParser) Parse(context.Context) (Heartbeats, error) {
	return nil, errors.New("failed")
}

func (errorParser) Name() string {
	return "error"
}

func TestParseHeartbeats_Panic(t *testing.T) {
	heartbeats, err := parseHeartbeats(t.Context(), panicParser{}, nil)

	require.Error(t, err)
	assert.Nil(t, heartbeats)
	assert.Contains(t, err.Error(), "panicked: OOM")
}

func TestParseHeartbeats_PanicReportsDiagnostic(t *testing.T) {
	var report aiPanicReport

	heartbeats, err := parseHeartbeats(t.Context(), panicParser{}, func(r aiPanicReport) {
		report = r
	})

	require.Error(t, err)
	assert.Nil(t, heartbeats)
	assert.Equal(t, "panic", report.ParserName)
	assert.Equal(t, "OOM", report.Recovered)
	assert.Contains(t, report.Stack, "panicParser.Parse")
}

func TestParseHeartbeats_Success(t *testing.T) {
	heartbeats, err := parseHeartbeats(t.Context(), heartbeatsParser{}, nil)

	require.NoError(t, err)
	assert.Len(t, heartbeats, 2)
}

func TestParseHeartbeats_Error(t *testing.T) {
	heartbeats, err := parseHeartbeats(t.Context(), errorParser{}, nil)

	require.EqualError(t, err, "failed")
	assert.Nil(t, heartbeats)
}
