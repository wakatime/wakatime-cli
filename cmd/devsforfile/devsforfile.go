package devsforfile

import (
	"fmt"
	"strings"

	"github.com/wakatime/wakatime-cli/cmd/api"
	paramscmd "github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/pkg/apikey"
	"github.com/wakatime/wakatime-cli/pkg/devsforfile"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/filter"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"

	"github.com/spf13/viper"
)

// Run executes the devs-for-file command.
func Run(v *viper.Viper) (int, error) {
	output, err := DevsForFile(v)
	if err != nil {
		if errwaka, ok := err.(wakaerror.Error); ok {
			return errwaka.ExitCode(), fmt.Errorf("devs-for-file fetch failed: %s", errwaka.Message())
		}

		return exitcode.ErrGeneric, fmt.Errorf(
			"devs-for-file fetch failed: %s",
			err,
		)
	}

	log.Debugln("successfully fetched devs-for-file for status bar")
	fmt.Println(output)

	return exitcode.Success, nil
}

// DevsForFile returns the top devs for a file.
func DevsForFile(v *viper.Viper) (string, error) {
	params, err := paramscmd.Load(v)
	if err != nil {
		return "", fmt.Errorf("failed to load command parameters: %w", err)
	}

	userAgent := heartbeat.UserAgent(params.API.Plugin)

	heartbeats := []heartbeat.Heartbeat{
		heartbeat.New(
			params.Heartbeat.Project.BranchAlternate,
			params.Heartbeat.Category,
			params.Heartbeat.CursorPosition,
			params.Heartbeat.Entity,
			params.Heartbeat.EntityType,
			params.Heartbeat.IsUnsavedEntity,
			params.Heartbeat.IsWrite,
			params.Heartbeat.Language,
			params.Heartbeat.LanguageAlternate,
			params.Heartbeat.LineNumber,
			params.Heartbeat.LinesInFile,
			params.Heartbeat.LocalFile,
			params.Heartbeat.Project.Alternate,
			params.Heartbeat.Project.Override,
			params.Heartbeat.Sanitize.ProjectPathOverride,
			params.Heartbeat.Time,
			userAgent,
		),
	}

	handleOpts := initHandleOptions(params)

	sender := mockSender{
		SendHeartbeatsFn: func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			return []heartbeat.Result{
				{
					Status: 201,
					Heartbeat: heartbeat.Heartbeat{
						Entity:           hh[0].Entity,
						Project:          hh[0].Project,
						ProjectRootCount: hh[0].ProjectRootCount,
					},
				},
			}, nil
		},
	}

	handle := heartbeat.NewHandle(&sender, handleOpts...)
	results, err := handle(heartbeats)
	if err != nil {
		return "", err
	}

	apiClient, err := api.NewClient(params.API)
	if err != nil {
		return "", fmt.Errorf("failed to initialize api client: %w", err)
	}

	if len(results) == 1 {
		result := results[0]

		if len(result.Errors) > 0 {
			log.Warnln(strings.Join(result.Errors, " "))
		}

		s, err := apiClient.DevsForFile(result.Heartbeat)
		if err != nil {
			return "", fmt.Errorf("failed fetching devs-for-file from api: %w", err)
		}

		output, err := devsforfile.RenderDevsForFile(s, params.StatusBar.Output)
		if err != nil {
			return "", fmt.Errorf("failed generating devs-for-file output: %s", err)
		}

		return output, nil
	}

	return "", fmt.Errorf("no heartbeat left")
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

type mockSender struct {
	SendHeartbeatsFn        func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error)
	SendHeartbeatsFnInvoked bool
}

func (m *mockSender) SendHeartbeats(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
	m.SendHeartbeatsFnInvoked = true
	return m.SendHeartbeatsFn(hh)
}
