package offline

import (
	"context"
	"errors"
	"fmt"

	"github.com/wakatime/wakatime-cli/cmd/handler"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/params"

	"github.com/spf13/viper"
)

// SaveHeartbeats saves heartbeats to the offline db without trying to send to the API.
// Used when we have more heartbeats than `offline.SendLimit`, when we couldn't send
// heartbeats to the API, or the API returned an auth error.
func SaveHeartbeats(ctx context.Context, v *viper.Viper, heartbeats []heartbeat.Heartbeat, queueFilepath string) error {
	params, err := LoadParams(ctx, v)
	if err != nil {
		return fmt.Errorf("failed to load command parameters: %w", err)
	}

	logger := log.Extract(ctx)

	setLogFields(ctx, params)
	logger.Debugf("params: %s", params)

	if params.Offline.Disabled {
		return errors.New("saving to offline db disabled")
	}

	if heartbeats == nil {
		// We're not saving surplus extra heartbeats, so save
		// main heartbeat and all extra heartbeats to offline db
		heartbeats = buildHeartbeats(ctx, params)
	}

	handleOpts := initHandleOptions()
	sender := heartbeat.NewHandle(Noop{}, offline.WithQueue(queueFilepath))
	handle := handler.New(v, handler.Config{
		Params:       params,
		ParamsLoader: LoadParams,
		Opts:         handleOpts,
	})(sender)

	_, _ = handle(ctx, heartbeats)

	return nil
}

// LoadParams loads params from viper.Viper instance. Returns ErrAuth
// if failed to retrieve api key.
func LoadParams(ctx context.Context, v *viper.Viper) (params.Params, error) {
	logger := log.Extract(ctx)

	paramAPI, err := params.LoadAPIParams(ctx, v)
	if err != nil {
		logger.Warnf("failed to load API parameters: %s", err)
	}

	paramHeartbeat, err := params.LoadHeartbeatParams(ctx, v)
	if err != nil {
		return params.Params{}, fmt.Errorf("failed to load heartbeat parameters: %s", err)
	}

	return params.Params{
		API:       paramAPI,
		Heartbeat: paramHeartbeat,
		Offline:   params.LoadOfflineParams(ctx, v),
	}, nil
}

func buildHeartbeats(ctx context.Context, params params.Params) []heartbeat.Heartbeat {
	heartbeats := []heartbeat.Heartbeat{}

	userAgent := heartbeat.UserAgent(ctx, params.API.Plugin)

	heartbeats = append(heartbeats, heartbeat.New(
		params.Heartbeat.Project.BranchAlternate,
		params.Heartbeat.Category,
		params.Heartbeat.CursorPosition,
		params.Heartbeat.Entity,
		params.Heartbeat.EntityType,
		params.Heartbeat.IsUnsavedEntity,
		params.Heartbeat.IsWrite,
		params.Heartbeat.Language,
		params.Heartbeat.LanguageAlternate,
		params.Heartbeat.LineAdditions,
		params.Heartbeat.LineDeletions,
		params.Heartbeat.LineNumber,
		params.Heartbeat.LinesInFile,
		params.Heartbeat.LocalFile,
		params.Heartbeat.Project.Alternate,
		params.Heartbeat.Project.ProjectFromGitRemote,
		params.Heartbeat.Project.Override,
		params.Heartbeat.Sanitize.ProjectPathOverride,
		params.Heartbeat.Time,
		userAgent,
	))

	if len(params.Heartbeat.ExtraHeartbeats) > 0 {
		logger := log.Extract(ctx)
		logger.Debugf("include %d extra heartbeat(s) from stdin", len(params.Heartbeat.ExtraHeartbeats))

		for _, h := range params.Heartbeat.ExtraHeartbeats {
			h.UserAgent = userAgent

			heartbeats = append(heartbeats, h)
		}
	}

	return heartbeats
}

func initHandleOptions() []handler.Preprocessor {
	return []handler.Preprocessor{
		handler.WithFormatting(),
		handler.WithEntityModifier(),
		handler.WithHeartbeatFiltering(),
		handler.WithRemoteDetection(),
		handler.WithFileStatsDetection(),
		handler.WithLanguageDetection(),
		handler.WithDependencyDetection(),
		handler.WithProjectDetection(),
		handler.WithProjectFiltering(),
		handler.WithHeartbeatSanitization(),
		handler.WithRemoteCleanup(),
		handler.WithLengthValidator(),
	}
}

func setLogFields(ctx context.Context, params params.Params) {
	log.AddField(ctx, "file", params.Heartbeat.Entity)
	log.AddField(ctx, "time", params.Heartbeat.Time)

	if params.API.Plugin != "" {
		log.AddField(ctx, "plugin", params.API.Plugin)
	}

	if params.Heartbeat.LineNumber != nil {
		log.AddField(ctx, "lineno", params.Heartbeat.LineNumber)
	}

	if params.Heartbeat.IsWrite != nil {
		log.AddField(ctx, "is_write", params.Heartbeat.IsWrite)
	}
}

// Noop is a noop api client, used by offline.SaveHeartbeats.
type Noop struct{}

// SendHeartbeats always returns an error.
func (Noop) SendHeartbeats(_ context.Context, _ []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
	return nil, api.Err{
		Err: errors.New("skip sending heartbeats and only save to offline db"),
	}
}
