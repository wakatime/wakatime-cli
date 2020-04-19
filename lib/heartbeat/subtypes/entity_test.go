package subtypes_test

import (
	"encoding/json"
	"testing"

	"github.com/alanhamlett/wakatime-cli/lib/heartbeat/subtypes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var typeTests = map[string]subtypes.EntityType{
	"file":   subtypes.FileType,
	"domain": subtypes.DomainType,
	"app":    subtypes.AppType,
}

func TestEntityType_UnmarshalJSON(t *testing.T) {
	for value, entityType := range typeTests {
		t.Run(value, func(t *testing.T) {
			var et subtypes.EntityType
			require.NoError(t, json.Unmarshal([]byte(`"`+value+`"`), &et))

			assert.Equal(t, entityType, et)
		})
	}
}

func TestEntityType_UnmarshalJSON_Invalid(t *testing.T) {
	var et subtypes.EntityType
	require.Error(t, json.Unmarshal([]byte(`"invalid"`), &et))
}

func TestEntityType_MarshalJSON(t *testing.T) {
	for value, entityType := range typeTests {
		t.Run(value, func(t *testing.T) {
			data, err := json.Marshal(entityType)
			require.NoError(t, err)
			assert.JSONEq(t, `"`+value+`"`, string(data))
		})
	}
}

func TestEntityType_String(t *testing.T) {
	for value, entityType := range typeTests {
		t.Run(value, func(t *testing.T) {
			s := entityType.String()
			assert.Equal(t, value, s)
		})
	}
}
