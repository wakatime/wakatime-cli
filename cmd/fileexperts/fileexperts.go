package fileexperts

import (
	"fmt"

	apicmd "github.com/wakatime/wakatime-cli/cmd/api"
	paramscmd "github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/pkg/apikey"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/fileexperts"
	"github.com/wakatime/wakatime-cli/pkg/filter"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"

	"github.com/spf13/viper"
)

// Params contains file-expert command parameters.
type Params struct {
	Entity string
	Params paramscmd.Params
}

// Run executes the file-experts command.
func Run(v *viper.Viper) (int, error) {
	output, err := FileExperts(v)
	if err != nil {
		if errwaka, ok := err.(wakaerror.Error); ok {
			return errwaka.ExitCode(), fmt.Errorf("file experts fetch failed: %s", errwaka.Message())
		}

		return exitcode.ErrGeneric, fmt.Errorf(
			"file experts fetch failed: %s",
			err,
		)
	}

	log.Debugln("successfully fetched file experts")
	fmt.Println(output)

	return exitcode.Success, nil
}

// FileExperts returns a rendered file experts of todays coding activity.
func FileExperts(v *viper.Viper) (string, error) {
	params, err := LoadParams(v)
	if err != nil {
		return "", err
	}

	handleOpts := initHandleOptions(params.Params)

	apiClient, err := apicmd.NewClientWithoutAuth(params.Params.API)
	if err != nil {
		return "", fmt.Errorf("failed to initialize api client: %w", err)
	}

	handle := fileexperts.NewHandle(apiClient, handleOpts...)

	results, err := handle([]heartbeat.Heartbeat{{Entity: params.Entity}})
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", nil
	}

	output, err := fileexperts.RenderFileExperts(
		results[0].FileExpert.(*fileexperts.FileExperts),
		params.Params.StatusBar.Output,
	)
	if err != nil {
		return "", fmt.Errorf("failed generating fileexpert output: %s", err)
	}

	return output, nil
}

// LoadParams loads file-expert config params from viper.Viper instance. Returns ErrAuth
// if failed to retrieve api key.
func LoadParams(v *viper.Viper) (Params, error) {
	if v == nil {
		return Params{}, fmt.Errorf("viper instance is nil")
	}

	params, err := paramscmd.Load(v)
	if err != nil {
		return Params{}, fmt.Errorf("failed to load command parameters: %w", err)
	}

	return Params{
		Entity: params.Heartbeat.Entity,
		Params: params,
	}, nil
}

func initHandleOptions(params paramscmd.Params) []heartbeat.HandleOption {
	return []heartbeat.HandleOption{
		heartbeat.WithFormatting(),
		heartbeat.WithEntityModifer(),
		filter.WithFiltering(filter.Config{
			Exclude:                    params.Heartbeat.Filter.Exclude,
			Include:                    params.Heartbeat.Filter.Include,
			IncludeOnlyWithProjectFile: params.Heartbeat.Filter.IncludeOnlyWithProjectFile,
		}),
		apikey.WithReplacing(apikey.Config{
			DefaultAPIKey: params.API.Key,
			MapPatterns:   params.API.KeyPatterns,
		}),
		project.WithDetection(project.Config{
			HideProjectNames: params.Heartbeat.Sanitize.HideProjectNames,
			MapPatterns:      params.Heartbeat.Project.MapPatterns,
			Submodule: project.Submodule{
				DisabledPatterns: params.Heartbeat.Project.SubmodulesDisabled,
				MapPatterns:      params.Heartbeat.Project.SubmoduleMapPatterns,
			},
		}),
		project.WithFiltering(project.FilterConfig{
			ExcludeUnknownProject: params.Heartbeat.Filter.ExcludeUnknownProject,
		}),
		heartbeat.WithSanitization(heartbeat.SanitizeConfig{
			BranchPatterns:    params.Heartbeat.Sanitize.HideBranchNames,
			FilePatterns:      params.Heartbeat.Sanitize.HideFileNames,
			HideProjectFolder: params.Heartbeat.Sanitize.HideProjectFolder,
			ProjectPatterns:   params.Heartbeat.Sanitize.HideProjectNames,
		}),
		filter.WithLengthValidator(),
	}
}
