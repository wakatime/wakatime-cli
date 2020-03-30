package heartbeat_test

import (
	"testing"

	"github.com/alanhamlett/wakatime-cli/lib/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var typeTests = map[string]heartbeat.EntityType{
	"file":   heartbeat.FileType,
	"domain": heartbeat.DomainType,
	"app":    heartbeat.AppType,
}

func TestEntityType_UnmarshalJSON(t *testing.T) {
	for value, entityType := range typeTests {
		t.Run(value, func(t *testing.T) {
			var et heartbeat.EntityType
			err := et.UnmarshalJSON([]byte(value))
			require.NoError(t, err)

			assert.Equal(t, entityType, et)
		})
	}
}

func TestEntityType_UnmarshalJSON_Invalid(t *testing.T) {
	var et heartbeat.EntityType
	err := et.UnmarshalJSON([]byte("invalid"))
	require.Error(t, err)
}

func TestEntityType_MarshalJSON(t *testing.T) {
	for value, entityType := range typeTests {
		t.Run(value, func(t *testing.T) {
			val, err := entityType.MarshalJSON()
			require.NoError(t, err)

			assert.Equal(t, value, string(val))
		})
	}
}
