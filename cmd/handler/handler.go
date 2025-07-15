package handler

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/wakatime/wakatime-cli/pkg/file"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
	paramspkg "github.com/wakatime/wakatime-cli/pkg/params"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/spf13/viper"
)

const projectConfigFileName = ".wakatime"

// Config contains the configuration for the heartbeat handler.
type Config struct {
	Params       paramspkg.Params
	ParamsLoader func(context.Context, *viper.Viper) (paramspkg.Params, error)
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
	paramsLoader func(context.Context, *viper.Viper) (paramspkg.Params, error),
) (paramspkg.Params, bool, error) {
	if entityType != heartbeat.FileType || isUnsavedEntity {
		return paramspkg.Params{}, false, nil
	}

	fp, ok := file.Find(ctx, filepath.Dir(entity), projectConfigFileName)
	if !ok {
		return paramspkg.Params{}, false, nil
	}

	vproj, err := vipertools.New()
	if err != nil {
		return paramspkg.Params{}, false, fmt.Errorf("failed to create viper instance: %w", err)
	}

	// copy only settings from the main viper instance to the project-level viper instance
	vipertools.CopyOnlySettings(v, vproj)

	// load project-level configuration file into viper instance
	if err := ini.ReadInConfig(vproj, fp); err != nil {
		return paramspkg.Params{}, false, fmt.Errorf("failed to load project-level configuration file: %s", err)
	}

	// load parameters from viper instance
	params, err := paramsLoader(ctx, vproj)
	if err != nil {
		return paramspkg.Params{}, false, fmt.Errorf("failed to load project-level parameters: %w", err)
	}

	return params, true, nil
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
