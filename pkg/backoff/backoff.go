package backoff

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/spf13/viper"
)

const (
	// maxBackoff sets the maximum seconds we will rate limit before retrying to send.
	maxBackoffSecs = 3600
	// factor is the total seconds to be multiplied by.
	factor = 15
)

// Config defines backoff data.
type Config struct {
	// At is the time when the first failure happened.
	At time.Time
	// Retries is the number of attempts to connect.
	Retries int
	// V is an instance of Viper.
	V *viper.Viper
	// HasProxy is true when using a proxy
	HasProxy bool
}

// WithBackoff initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to prevent trying to send
// a heartbeat when the api is unresponsive.
func WithBackoff(config Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(ctx context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			logger := log.Extract(ctx)
			logger.Debugln("execute heartbeat backoff algorithm")

			if shouldBackoff(ctx, config.Retries, config.At) {
				if config.HasProxy {
					return nil, api.ErrBackoff{Err: errors.New("won't send heartbeat due to backoff with proxy")}
				}

				return nil, api.ErrBackoff{Err: errors.New("won't send heartbeat due to backoff without proxy")}
			}

			results, err := next(ctx, hh)
			if err != nil {
				// error response, increment backoff
				if updateErr := updateBackoffSettings(ctx, config.V, config.Retries+1, time.Now()); updateErr != nil {
					logger.Warnf("failed to update backoff settings: %s", updateErr)
				}

				return nil, err
			}

			// success response, reset backoff
			if config.Retries > 0 || !config.At.IsZero() {
				if resetErr := updateBackoffSettings(ctx, config.V, 0, time.Time{}); resetErr != nil {
					logger.Warnf("failed to reset backoff settings: %s", resetErr)
				}
			}

			return results, nil
		}
	}
}

// shouldBackoff returns true if we should save heartbeats directly to offline
// database and skip sending to API due to rate limiting from too many recent
// networking errors.
func shouldBackoff(ctx context.Context, retries int, at time.Time) bool {
	if retries < 1 || at.IsZero() {
		return false
	}

	logger := log.Extract(ctx)

	backoffSeconds := float64(factor) * math.Pow(2, float64(retries))

	duration := time.Duration(backoffSeconds) * time.Second

	if backoffSeconds > maxBackoffSecs {
		logger.Debugf(
			"exponential backoff tried %d times since %s, will reset because reached %s max backoff",
			retries,
			at.Format(ini.DateFormat),
			duration.String(),
		)

		return false
	}

	if at.Add(duration).Before(time.Now()) {
		return false
	}

	logger.Debugf(
		"exponential backoff tried %d times since %s, will retry again after %s",
		retries,
		at.Format(ini.DateFormat),
		time.Now().Add(duration).Format(ini.DateFormat),
	)

	return true
}

func updateBackoffSettings(ctx context.Context, v *viper.Viper, retries int, at time.Time) error {
	w, err := ini.NewWriter(ctx, v, ini.InternalFilePath)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %s", err)
	}

	keyValue := map[string]string{
		"backoff_retries": strconv.Itoa(retries),
		"backoff_at":      "",
	}

	if !at.IsZero() {
		keyValue["backoff_at"] = at.Format(ini.DateFormat)
	} else {
		keyValue["backoff_at"] = ""
	}

	if err := w.Write(ctx, "internal", keyValue); err != nil {
		return fmt.Errorf("failed to write to internal config file: %s", err)
	}

	return nil
}
