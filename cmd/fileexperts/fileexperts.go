package fileexperts

import (
	"context"
	"fmt"

	apicmd "github.com/wakatime/wakatime-cli/cmd/api"
	"github.com/wakatime/wakatime-cli/cmd/handler"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/fileexperts"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/params"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"

	"github.com/spf13/viper"
)

// Run executes the file-experts command.
func Run(ctx context.Context, v *viper.Viper) (int, error) {
	output, err := FileExperts(ctx, v)
	if err != nil {
		if errwaka, ok := err.(wakaerror.Error); ok {
			return errwaka.ExitCode(), fmt.Errorf("file experts fetch failed: %s", errwaka.Message())
		}

		return exitcode.ErrGeneric, fmt.Errorf(
			"file experts fetch failed: %s",
			err,
		)
	}

	logger := log.Extract(ctx)
	logger.Debugln("successfully fetched file experts")

	fmt.Println(output)

	return exitcode.Success, nil
}

// FileExperts returns a rendered file experts of todays coding activity.
func FileExperts(ctx context.Context, v *viper.Viper) (string, error) {
	params, err := LoadParams(ctx, v)
	if err != nil {
		return "", fmt.Errorf("failed to load command parameters: %w", err)
	}

	logger := log.Extract(ctx)

	setLogFields(ctx, params)
	logger.Debugf("params: %s", params)

	handleOpts := initHandleOptions()

	apiClient, err := apicmd.NewClientWithoutAuth(ctx, params.API)
	if err != nil {
		return "", fmt.Errorf("failed to initialize api client: %w", err)
	}

	sender := fileexperts.NewHandle(apiClient)
	handle := handler.New(v, handler.Config{
		Params:       params,
		ParamsLoader: LoadParams,
		Opts:         handleOpts,
	})(sender)

	results, err := handle(ctx, []heartbeat.Heartbeat{{Entity: params.Heartbeat.Entity}})
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", nil
	}

	output, err := fileexperts.RenderFileExperts(
		results[0].FileExpert.(*fileexperts.FileExperts),
		params.StatusBar.Output,
	)
	if err != nil {
		return "", fmt.Errorf("failed generating fileexpert output: %s", err)
	}

	return output, nil
}

// LoadParams loads file-expert config params from viper.Viper instance. Returns ErrAuth
// if failed to retrieve api key.
func LoadParams(ctx context.Context, v *viper.Viper) (params.Params, error) {
	if v == nil {
		return params.Params{}, fmt.Errorf("viper instance unset")
	}

	heartbeatParams, err := params.LoadHeartbeatParams(ctx, v)
	if err != nil {
		return params.Params{}, fmt.Errorf("failed to load heartbeat params: %s", err)
	}

	apiParams, err := params.LoadAPIParams(ctx, v)
	if err != nil {
		return params.Params{}, fmt.Errorf("failed to load API parameters: %w", err)
	}

	statusBarParams, err := params.LoadStatusBarParams(v)
	if err != nil {
		return params.Params{}, fmt.Errorf("failed to load status bar params: %w", err)
	}

	return params.Params{
		API:       apiParams,
		Heartbeat: heartbeatParams,
		StatusBar: statusBarParams,
	}, nil
}

func initHandleOptions() []handler.Preprocessor {
	return []handler.Preprocessor{
		handler.WithFormatting(),
		handler.WithEntityModifier(),
		handler.WithHeartbeatFiltering(),
		handler.WithRemoteDetection(),
		handler.WithAPIKeyReplacing(),
		handler.WithProjectDetection(),
		handler.WithProjectFiltering(),
		handler.WithHeartbeatSanitization(),
		handler.WithFileExpertsValidation(),
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
