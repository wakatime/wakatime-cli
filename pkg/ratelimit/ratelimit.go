package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/ini"

	"github.com/spf13/viper"
)

// Params contains the parameters used to determine if heartbeats should be rate limited.
type Params struct {
	Disabled   bool
	LastSentAt time.Time
	Timeout    time.Duration
}

// IsRateLimited determines if we should send heartbeats to the API or save to the offline db.
func IsRateLimited(params Params) bool {
	if params.Disabled {
		return false
	}

	if params.Timeout == 0 {
		return false
	}

	if params.LastSentAt.IsZero() {
		return false
	}

	return time.Since(params.LastSentAt) < params.Timeout
}

// Reset updates the internal.heartbeats_last_sent_at timestamp.
func Reset(ctx context.Context, v *viper.Viper) error {
	w, err := ini.NewWriter(ctx, v, ini.InternalFilePath)
	if err != nil {
		return fmt.Errorf("failed to parse internal config file: %s", err)
	}

	keyValue := map[string]string{
		"heartbeats_last_sent_at": time.Now().Format(ini.DateFormat),
	}

	if err := w.Write(ctx, "internal", keyValue); err != nil {
		return fmt.Errorf("failed to write to internal config file: %s", err)
	}

	return nil
}
