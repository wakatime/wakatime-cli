package heartbeat_test

import (
	"encoding/json"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func typeTests() map[string]heartbeat.EntityType {
	return map[string]heartbeat.EntityType{
		"file":   heartbeat.FileType,
		"domain": heartbeat.DomainType,
		"app":    heartbeat.AppType,
	}
}

func TestParseEntityType(t *testing.T) {
	for value, entityType := range typeTests() {
		t.Run(value, func(t *testing.T) {
			parsed, err := heartbeat.ParseEntityType(value)
			require.NoError(t, err)

			assert.Equal(t, entityType, parsed)
		})
	}
}

func TestParseEntityType_Invalid(t *testing.T) {
	_, err := heartbeat.ParseEntityType("invalid")
	require.Error(t, err)
}

func TestEntityType_UnmarshalJSON(t *testing.T) {
	for value, entityType := range typeTests() {
		t.Run(value, func(t *testing.T) {
			var et heartbeat.EntityType
			require.NoError(t, json.Unmarshal([]byte(`"`+value+`"`), &et))

			assert.Equal(t, entityType, et)
		})
	}
}

func TestEntityType_UnmarshalJSON_Invalid(t *testing.T) {
	var et heartbeat.EntityType

	require.Error(t, json.Unmarshal([]byte(`"invalid"`), &et))
}

func TestEntityType_MarshalJSON(t *testing.T) {
	for value, entityType := range typeTests() {
		t.Run(value, func(t *testing.T) {
			data, err := json.Marshal(entityType)
			require.NoError(t, err)
			assert.JSONEq(t, `"`+value+`"`, string(data))
		})
	}
}

func TestEntityType_String(t *testing.T) {
	for value, entityType := range typeTests() {
		t.Run(value, func(t *testing.T) {
			s := entityType.String()
			assert.Equal(t, value, s)
		})
	}
}
