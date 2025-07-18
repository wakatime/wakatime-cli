package handler

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/wakatime/wakatime-cli/pkg/file"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/params"

	"github.com/spf13/viper"
)

const projectConfigFileName = ".wakatime"

// Config contains the configuration for the heartbeat handler.
type Config struct {
	Params       params.Params
	ParamsLoader func(context.Context, *viper.Viper, params.FlagReadOrder) (params.Params, error)
	Opts         []Preprocessor
}

// New creates a new heartbeat.HandleOption that processes heartbeats
// with project-level configuration files and applies the provided options.
func New(
	v *viper.Viper,
	cfg Config,
) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(ctx context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			var heartbeats []heartbeat.Heartbeat

			logger := *log.Extract(ctx) // make it a shallow copy to avoid modifying the original logger

			for _, h := range hh {
				// shallow copy the params to avoid modifying the original
				effectiveParams := cfg.Params

				p, ok, err := findProjectConfigFile(ctx, v, h.Entity, h.EntityType, h.IsUnsavedEntity, cfg.ParamsLoader)
				if err != nil {
					return nil, fmt.Errorf("failed to find project configuration file: %w", err)
				}

				if ok {
					// if project-level configuration file is found, use it
					logger.Debugf("found project-level configuration file %s for entity %q", projectConfigFileName, h.Entity)
					logger.Debugf("project-level params: %s", p)

					effectiveParams = p
				}

				chain := heartbeat.NewHandle(noop{})

				for i := len(cfg.Opts) - 1; i >= 0; i-- {
					chain = cfg.Opts[i](effectiveParams)(chain)
				}

				res, err := chain(ctx, []heartbeat.Heartbeat{h})
				if err != nil {
					return nil, fmt.Errorf("failed to process heartbeat: %w", err)
				}

				for _, r := range res {
					heartbeats = append(heartbeats, r.Heartbeat)
				}
			}

			if len(heartbeats) == 0 {
				return []heartbeat.Result{}, nil
			}

			return next(ctx, heartbeats)
		}
	}
}

func findProjectConfigFile(
	ctx context.Context,
	v *viper.Viper,
	entity string,
	entityType heartbeat.EntityType,
	isUnsavedEntity bool,
	paramsLoader func(context.Context, *viper.Viper, params.FlagReadOrder) (params.Params, error),
) (params.Params, bool, error) {
	if entityType != heartbeat.FileType || isUnsavedEntity {
		return params.Params{}, false, nil
	}

	fp, ok := file.Find(ctx, filepath.Dir(entity), projectConfigFileName)
	if !ok {
		return params.Params{}, false, nil
	}

	// load project-level configuration file into viper instance
	if err := ini.ReadInConfig(v, fp); err != nil {
		return params.Params{}, false, fmt.Errorf("failed to load project-level configuration file: %s", err)
	}

	// load parameters from viper instance
	p, err := paramsLoader(ctx, v, params.FlagReadOrderProjectConfigPrecedence)
	if err != nil {
		return params.Params{}, false, fmt.Errorf("failed to load project-level parameters: %w", err)
	}

	return p, true, nil
}

// noop is a noop api client, used to format heartbeats without sending them.
type noop struct{}

// SendHeartbeats always returns results containing formatted heartbeats and no error.
func (noop) SendHeartbeats(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
	results := make([]heartbeat.Result, len(hh))

	for i := range hh {
		results[i] = heartbeat.Result{Heartbeat: hh[i]}
	}

	return results, nil
}
