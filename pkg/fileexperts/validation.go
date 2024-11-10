package fileexperts

import (
	"context"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// WithValidation initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to validate the heartbeat
// before sending it to the API.
func WithValidation() heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(ctx context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			logger := log.Extract(ctx)
			logger.Debugln("execute fileexperts validation")

			var filtered []heartbeat.Heartbeat

			for _, h := range hh {
				if !Validate(h) {
					logger.Debugf("missing required fields for fileexperts")
					continue
				}

				filtered = append(filtered, h)
			}

			return next(ctx, filtered)
		}
	}
}

// Validate validates if required fields are not empty.
func Validate(h heartbeat.Heartbeat) bool {
	if h.Entity == "" ||
		h.Project == nil || *h.Project == "" ||
		h.ProjectRootCount == nil || *h.ProjectRootCount == 0 {
		return false
	}

	return true
}
