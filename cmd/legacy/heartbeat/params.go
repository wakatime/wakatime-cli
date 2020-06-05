package heartbeat

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/spf13/viper"
)

// nolint
var apiKeyRegex = regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$")

// Params contains heartbeat command parameters.
type Params struct {
	APIKey     string
	APIUrl     string
	Category   heartbeat.Category
	Entity     string
	EntityType heartbeat.EntityType
	Hostname   string
	IsWrite    *bool
	Plugin     string
	Time       float64
	Timeout    time.Duration
}

// LoadParams loads heartbeat config params from viper.Viper instance. Returns ErrAuth
// if failed to retrieve api key.
func LoadParams(v *viper.Viper) (Params, error) {
	apiKey, ok := vipertools.FirstNonEmptyString(v, "key", "settings.api_key", "settings.apikey")
	if !ok {
		return Params{}, api.ErrAuth("failed to load api key")
	}

	if !apiKeyRegex.Match([]byte(apiKey)) {
		return Params{}, api.ErrAuth("invalid api key format")
	}

	apiURL := api.BaseURL
	if url, ok := vipertools.FirstNonEmptyString(v, "api-url", "apiurl", "settings.api_url"); ok {
		apiURL = url
	}

	var category heartbeat.Category

	if categoryStr := v.GetString("category"); categoryStr != "" {
		parsed, err := heartbeat.ParseCategory(categoryStr)
		if err != nil {
			return Params{}, fmt.Errorf("failed to parse category: %s", err)
		}

		category = parsed
	}

	entity, ok := vipertools.FirstNonEmptyString(v, "entity", "file")
	if !ok {
		return Params{}, errors.New("failed to retrieve entity")
	}

	var entityType heartbeat.EntityType

	if entityTypeStr := v.GetString("entity-type"); entityTypeStr != "" {
		parsed, err := heartbeat.ParseEntityType(entityTypeStr)
		if err != nil {
			return Params{}, fmt.Errorf("failed to parse entity type: %s", err)
		}

		entityType = parsed
	}

	var err error

	hostname, ok := vipertools.FirstNonEmptyString(v, "hostname", "settings.hostname")
	if !ok {
		hostname, err = os.Hostname()
		if err != nil {
			return Params{}, fmt.Errorf("failed to retrieve hostname from system: %s", err)
		}
	}

	var isWrite *bool
	if b := v.GetBool("write"); v.IsSet("write") {
		isWrite = heartbeat.Bool(b)
	}

	timeSecs := v.GetFloat64("time")
	if timeSecs == 0 {
		timeSecs = float64(time.Now().Unix())
	}

	var timeout time.Duration

	timeoutSecs, ok := vipertools.FirstNonEmptyInt(v, "timeout", "settings.timeout")
	if ok {
		timeout = time.Duration(timeoutSecs) * time.Second
	}

	return Params{
		APIKey:     apiKey,
		APIUrl:     apiURL,
		Category:   category,
		Entity:     entity,
		EntityType: entityType,
		Hostname:   hostname,
		IsWrite:    isWrite,
		Plugin:     v.GetString("plugin"),
		Time:       timeSecs,
		Timeout:    timeout,
	}, nil
}