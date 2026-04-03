package heartbeat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestShouldUseProjectConfig(t *testing.T) {
	assert.False(t, shouldUseProjectConfig(nil))
	assert.False(t, shouldUseProjectConfig([]heartbeat.Heartbeat{
		{Entity: "Claude session", EntityType: heartbeat.AppType},
	}))
	assert.False(t, shouldUseProjectConfig([]heartbeat.Heartbeat{
		{Entity: "untitled", EntityType: heartbeat.FileType, IsUnsavedEntity: true},
	}))
	assert.True(t, shouldUseProjectConfig([]heartbeat.Heartbeat{
		{Entity: "Claude session", EntityType: heartbeat.AppType},
		{Entity: "/tmp/main.go", EntityType: heartbeat.FileType},
	}))
}
