package backoff

import (
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
	// resetAfter sets the total seconds a backoff will last.
	resetAfter = 3600
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
}

// WithBackoff initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to prevent trying to send
// a heartbeat when the api is unresponsive.
func WithBackoff(config Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			log.Debugln("execute heartbeat backoff algorithm")

			should, reset := shouldBackoff(config.Retries, config.At)
			if should {
				return nil, api.ErrBackoff{Err: errors.New("won't send heartbeat due to backoff")}
			}

			if reset {
				config.Retries = 0
			}

			results, err := next(hh)
			if err != nil {
				log.Debugf("incrementing backoff due to error")

				// error response, increment backoff
				if updateErr := updateBackoffSettings(config.V, config.Retries+1, time.Now()); updateErr != nil {
					log.Warnf("failed to update backoff settings: %s", updateErr)
				}

				return nil, err
			}

			if reset || !config.At.IsZero() {
				// reset or success response, reset backoff
				if resetErr := updateBackoffSettings(config.V, 0, time.Time{}); resetErr != nil {
					log.Warnf("failed to reset backoff settings: %s", resetErr)
				}
			}

			return results, nil
		}
	}
}

// shouldBackoff returns true if the backoff should be applied. It also returns a boolean value
// indicating whether the backoff should be reset.
func shouldBackoff(retries int, at time.Time) (bool, bool) {
	if retries < 1 || at.IsZero() {
		return false, true
	}

	next := float64(factor) * math.Pow(2, float64(retries))
	if next > float64(resetAfter) {
		log.Debugf(
			"exponential backoff tried %d times since %s, it will reset due to %d seconds limit",
			retries,
			at.Format(time.Stamp),
			resetAfter,
		)

		return false, true
	}

	duration := time.Duration(next) * time.Second

	log.Debugf(
		"exponential backoff tried %d times since %s, will retry at %s",
		retries,
		at.Format(time.Stamp),
		at.Add(duration).Format(time.Stamp),
	)

	return time.Now().Before(at.Add(duration)), false
}

func updateBackoffSettings(v *viper.Viper, retries int, at time.Time) error {
	w, err := ini.NewWriter(v, ini.InternalFilePath)
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

	if err := w.Write("internal", keyValue); err != nil {
		return fmt.Errorf("failed to write to internal config file: %s", err)
	}

	return nil
}
