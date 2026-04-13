package offline

import (
	"context"
	"errors"
	"fmt"

	"github.com/wakatime/wakatime-cli/cmd/handler"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/filter"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/params"

	"github.com/spf13/viper"
)

// SaveHeartbeats saves heartbeats to the offline db without trying to send to the API.
// Used when we have more heartbeats than `offline.SendLimit`, when we couldn't send
// heartbeats to the API, or the API returned an auth error.
func SaveHeartbeats(
	ctx context.Context,
	v *viper.Viper,
	queueFilepath string,
	heartbeats []heartbeat.Heartbeat,
) error {
	params, err := loadParams(ctx, v, params.FlagReadOrderFlagPrecedence)
	if err != nil {
		return fmt.Errorf("failed to load command parameters: %w", err)
	}

	return saveHeartbeats(ctx, v, queueFilepath, heartbeats, params)
}

// SaveHeartbeatsWithParams saves heartbeats to the offline db using already loaded command params.
func SaveHeartbeatsWithParams(
	ctx context.Context,
	v *viper.Viper,
	queueFilepath string,
	heartbeats []heartbeat.Heartbeat,
	params params.Params,
) error {
	return saveHeartbeats(ctx, v, queueFilepath, heartbeats, params)
}

func saveHeartbeats(
	ctx context.Context,
	v *viper.Viper,
	queueFilepath string,
	heartbeats []heartbeat.Heartbeat,
	params params.Params,
) error {
	logger := log.Extract(ctx)

	setLogFields(ctx, params)
	logger.Debugf("params: %s", params)

	if params.Offline.Disabled {
		return errors.New("saving to offline db disabled")
	}

	if len(heartbeats) == 0 {
		return nil
	}

	if params.Offline.Disabled {
		return errors.New("saving to offline db disabled")
	}

	handleOpts := initHandleOptions()
	sender := heartbeat.NewHandle(Noop{}, filter.WithLengthValidator(), offline.WithQueue(queueFilepath))
	handle := handler.New(v, handler.Config{
		Params:       params,
		ParamsLoader: loadParams,
		Opts:         handleOpts,
	})(sender)

	_, _ = handle(ctx, heartbeats)

	return nil
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
		handler.WithCategoryDetection(),
		handler.WithProjectDetection(),
		handler.WithProjectFiltering(),
		handler.WithHeartbeatSanitization(),
		handler.WithRemoteCleanup(),
	}
}

func loadParams(
	ctx context.Context,
	v *viper.Viper,
	order params.FlagReadOrder,
) (params.Params, error) {
	if v == nil {
		return params.Params{}, errors.New("viper instance unset")
	}

	// ignore api param errors so we can still save heartbeats offline
	apiParams, _ := params.LoadAPIParams(ctx, v, order)

	aiParams, err := params.LoadAIParams(ctx, v, params.FlagReadOrderFlagPrecedence)
	if err != nil {
		return params.Params{}, fmt.Errorf("failed to load ai params: %w", err)
	}

	heartbeatParams, err := params.LoadHeartbeatParams(ctx, v, order)
	if err != nil {
		return params.Params{}, fmt.Errorf("failed to load heartbeat params: %s", err)
	}

	return params.Params{
		AI:        aiParams,
		API:       apiParams,
		Heartbeat: heartbeatParams,
		Offline:   params.LoadOfflineParams(ctx, v, order),
	}, nil
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
