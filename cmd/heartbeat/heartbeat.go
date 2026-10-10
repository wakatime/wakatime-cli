package heartbeat

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	apicmd "github.com/wakatime/wakatime-cli/cmd/api"
	"github.com/wakatime/wakatime-cli/cmd/handler"
	offlinecmd "github.com/wakatime/wakatime-cli/cmd/offline"
	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/backoff"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/file"
	"github.com/wakatime/wakatime-cli/pkg/filter"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	_ "github.com/wakatime/wakatime-cli/pkg/lexer" // force to load all lexers
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/params"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/ratelimit"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"

	"github.com/spf13/viper"
)

// Run executes the heartbeat command.
func Run(ctx context.Context, v *viper.Viper) (int, error) {
	logger := log.Extract(ctx)

	queueFilepath, err := offline.QueueFilepath(ctx, v)
	if err != nil {
		logger.Warnf("failed to load offline queue filepath: %s", err)
	}

	params, err := loadParams(ctx, v, params.FlagReadOrderFlagPrecedence)

	heartbeats := BuildHeartbeats(ctx, params.API.Plugin, params.Heartbeat)

	if err != nil {
		logger.Errorf("sending heartbeats failed: %s", err)

		if errSave := offlinecmd.SaveHeartbeats(ctx, v, queueFilepath, renderUserAgents(ctx, heartbeats)); errSave != nil {
			return exitcode.ErrConfigFileParse, fmt.Errorf("failed to save heartbeats to offline queue: %s", errSave)
		}

		return exitcode.ErrAuth, fmt.Errorf("failed to load heartbeat command parameters: %w", err)
	}

	err = SendHeartbeats(ctx, v, params, queueFilepath, heartbeats)
	if err != nil {
		var errauth api.ErrAuth

		// api.ErrAuth represents an error when parsing api key or timeout.
		// Save heartbeats to offline db when api.ErrAuth as it avoids losing heartbeats.
		if errors.As(err, &errauth) {
			if err := offlinecmd.SaveHeartbeats(ctx, v, queueFilepath, renderUserAgents(ctx, heartbeats)); err != nil {
				logger.Errorf("failed to save heartbeats to offline queue: %s", err)
			}

			return errauth.ExitCode(), fmt.Errorf("sending heartbeat(s) failed: %w", errauth)
		}

		var errwaka wakaerror.Error
		if errors.As(err, &errwaka) {
			return errwaka.ExitCode(), fmt.Errorf("sending heartbeat(s) failed: %w", err)
		}

		return exitcode.ErrGeneric, fmt.Errorf(
			"sending heartbeat(s) failed: %w",
			err,
		)
	}

	logger.Debugln("successfully sent heartbeat(s)")

	return exitcode.Success, nil
}

// SendHeartbeats sends heartbeats to the wakatime api and includes additional
// heartbeats from the offline queue, if available and offline sync is not
// explicitly disabled.
func SendHeartbeats(
	ctx context.Context,
	v *viper.Viper,
	params params.Params,
	queueFilepath string,
	heartbeats []heartbeat.Heartbeat,
) error {
	logger := log.Extract(ctx)

	setLogFields(ctx, params)
	logger.Debugf("params: %s", params)

	// Preserve normal IDE activity for projects excluded from AI tracking.
	// AI parsing can recategorize human heartbeats as "ai coding" or drop them
	// as duplicates; the later AI filter would then discard them. Excluded
	// humans must bypass AI parsing untouched, while excluded AI heartbeats
	// are still dropped later by the pipeline filter.
	included, excluded := splitAIExcludedHeartbeats(ctx, params, heartbeats)

	heartbeats, err := applyAIParsing(ctx, v, params, included)
	if err != nil {
		return err
	}

	heartbeats = append(excluded, heartbeats...)

	return sendPreparedHeartbeats(ctx, v, params, queueFilepath, heartbeats, true, loadParams)
}

func loadParams(
	ctx context.Context,
	v *viper.Viper,
	order params.FlagReadOrder,
) (params.Params, error) {
	var err error

	if v == nil {
		return params.Params{}, errors.New("viper instance unset")
	}

	heartbeatParams, err := params.LoadHeartbeatParams(ctx, v, order)
	if err != nil {
		return params.Params{}, fmt.Errorf("failed to load heartbeat params: %s", err)
	}

	apiParams, err := params.LoadAPIParams(ctx, v, order)
	if err != nil {
		err = fmt.Errorf("failed to load API parameters: %w", err)
	}

	return params.Params{
		AI:        heartbeatParams.AIParams,
		API:       apiParams,
		Heartbeat: heartbeatParams,
		Offline:   params.LoadOfflineParams(ctx, v, order),
	}, err
}

func buildHandle(ctx context.Context, v *viper.Viper, params params.Params, queueFilepath string) (heartbeat.Handle, error) {
	apiClient, err := apicmd.NewClientWithoutAuth(ctx, params.API)
	if err != nil {
		return nil, err
	}

	handleOpts := []heartbeat.HandleOption{
		filter.WithLengthValidator(),
	}

	if !params.Offline.Disabled {
		handleOpts = append(handleOpts, offline.WithQueue(queueFilepath))
	}

	handleOpts = append(handleOpts, backoff.WithBackoff(backoff.Config{
		V:        v,
		At:       params.API.BackoffAt,
		Retries:  params.API.BackoffRetries,
		HasProxy: params.API.ProxyURL != "",
	}))

	return heartbeat.NewHandle(apiClient, handleOpts...), nil
}

// BuildHeartbeats builds the command line heartbeat then appends any extra stdin heartbeats.
func BuildHeartbeats(ctx context.Context, plugin string, heartbeatParams params.Heartbeat) []heartbeat.Heartbeat {
	var heartbeats = make([]heartbeat.Heartbeat, 0, 1+len(heartbeatParams.ExtraHeartbeats))

	heartbeats = append(heartbeats, heartbeat.New(
		heartbeatParams.AILineChanges,
		heartbeatParams.Project.BranchAlternate,
		heartbeatParams.Category.String(),
		heartbeatParams.CursorPosition,
		heartbeatParams.Entity,
		heartbeatParams.EntityType,
		heartbeatParams.HumanLineChanges,
		heartbeatParams.IsUnsavedEntity,
		heartbeatParams.IsWrite,
		heartbeatParams.Language,
		heartbeatParams.LanguageAlternate,
		heartbeatParams.LineNumber,
		heartbeatParams.LinesInFile,
		heartbeatParams.LocalFile,
		heartbeatParams.Project.Alternate,
		heartbeatParams.Project.ProjectFromGitRemote,
		heartbeatParams.Project.Override,
		heartbeatParams.Sanitize.ProjectPathOverride,
		heartbeatParams.Time,
		plugin,
	))

	if len(heartbeatParams.ExtraHeartbeats) > 0 {
		logger := log.Extract(ctx)
		logger.Debugf("include %d extra heartbeat(s) from stdin", len(heartbeatParams.ExtraHeartbeats))

		for _, h := range heartbeatParams.ExtraHeartbeats {
			h.UserAgent = plugin

			heartbeats = append(heartbeats, h)
		}
	}

	return heartbeats
}

func renderUserAgents(ctx context.Context, heartbeats []heartbeat.Heartbeat) []heartbeat.Heartbeat {
	for i := range heartbeats {
		heartbeats[i].UserAgent = heartbeat.UserAgent(ctx, heartbeats[i].UserAgent)
	}

	return heartbeats
}

func initHandleOptions() []handler.Preprocessor {
	return []handler.Preprocessor{
		handler.WithFormatting(),
		handler.WithEntityModifier(),
		handler.WithHeartbeatFiltering(),
		handler.WithRemoteDetection(),
		handler.WithAPIKeyReplacing(),
		handler.WithFileStatsDetection(),
		handler.WithLanguageDetection(),
		handler.WithDependencyDetection(),
		handler.WithCategoryDetection(),
		handler.WithProjectDetection(),
		handler.WithProjectFiltering(),
		handler.WithAIFiltering(),
		handler.WithHeartbeatSanitization(),
		handler.WithRemoteCleanup(),
	}
}

// splitAIExcludedHeartbeats separates humans belonging to AI-excluded projects
// before AI parsing runs. AI parsing can recategorize human heartbeats as
// "ai coding" or drop them as duplicates, after which the pipeline AI filter
// would discard normal IDE activity. Excluded heartbeats bypass AI parsing
// untouched (normal coding is still logged); excluded AI heartbeats generated
// from transcripts are still dropped later by the pipeline filter after
// project detection.
func splitAIExcludedHeartbeats(
	ctx context.Context,
	params params.Params,
	hh []heartbeat.Heartbeat,
) (included, excluded []heartbeat.Heartbeat) {
	if len(hh) == 0 {
		return hh, nil
	}

	// Fast path: project-level .wakatime files can still disable AI even when
	// no global exclude patterns are configured. Check files first without
	// project detection.
	var remaining []heartbeat.Heartbeat

	for _, h := range hh {
		if projectFileDisablesAI(ctx, h) {
			excluded = append(excluded, h)
			continue
		}

		remaining = append(remaining, h)
	}

	if len(params.AI.ExcludeProjects) == 0 {
		return remaining, excluded
	}

	// Resolve project names so exclude patterns match the same names the
	// pipeline filter sees after project detection.
	resolved := resolveProjectsForExclusion(ctx, params, remaining)

	for _, h := range resolved {
		if h.Project != nil {
			matched := false

			for _, pattern := range params.AI.ExcludeProjects {
				if pattern.MatchString(ctx, *h.Project) {
					matched = true
					break
				}
			}

			if matched {
				excluded = append(excluded, h)
				continue
			}
		}

		included = append(included, h)
	}

	return included, excluded
}

// resolveProjectsForExclusion runs project detection to fill Project/Branch/
// ProjectPath, so AI exclusions match resolved names before AI parsing.
// It reuses the same detection config as the pipeline; running detection
// twice is idempotent.
func resolveProjectsForExclusion(
	ctx context.Context,
	params params.Params,
	hh []heartbeat.Heartbeat,
) []heartbeat.Heartbeat {
	if len(hh) == 0 {
		return hh
	}

	detect := project.WithDetection(project.Config{
		HideProjectNames:     params.Heartbeat.Sanitize.HideProjectNames,
		MapPatterns:          params.Heartbeat.Project.MapPatterns,
		ProjectFromGitRemote: params.Heartbeat.Project.ProjectFromGitRemote,
		Submodule: project.Submodule{
			DisabledPatterns: params.Heartbeat.Project.SubmodulesDisabled,
			MapPatterns:      params.Heartbeat.Project.SubmoduleMapPatterns,
		},
	})(func(_ context.Context, resolved []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		results := make([]heartbeat.Result, len(resolved))
		for i := range resolved {
			results[i] = heartbeat.Result{Heartbeat: resolved[i]}
		}

		return results, nil
	})

	results, err := detect(ctx, hh)
	if err != nil {
		return hh
	}

	resolved := make([]heartbeat.Heartbeat, 0, len(results))
	for _, r := range results {
		resolved = append(resolved, r.Heartbeat)
	}

	return resolved
}

// projectFileDisablesAI reports whether the project-level .wakatime file for
// the heartbeat's entity disables AI tracking. It mirrors handler's project
// config lookup but read-only, so the shared viper instance is not mutated
// before AI parsing.
func projectFileDisablesAI(ctx context.Context, h heartbeat.Heartbeat) bool {
	if h.EntityType != heartbeat.FileType || h.IsUnsavedEntity {
		return false
	}

	fp, ok := file.Find(ctx, filepath.Dir(h.Entity), ".wakatime")
	if !ok {
		return false
	}

	vv := viper.New()
	if err := ini.ReadInConfig(vv, fp); err != nil {
		return false
	}

	return vv.GetBool("settings.sync_ai_disabled") ||
		vv.GetBool("sync-ai-disabled") ||
		vv.GetBool("sync-ai-disable")
}

func applyAIParsing(
	ctx context.Context,
	v *viper.Viper,
	params params.Params,
	heartbeats []heartbeat.Heartbeat,
) ([]heartbeat.Heartbeat, error) {
	handle := ai.WithAISync(ai.Config{
		SyncDisabled: params.AI.SyncDisabled,
		Plugin:       params.API.Plugin,
		Project:      params.Heartbeat.Project,
		Sanitize:     params.Heartbeat.Sanitize,
		V:            v,
	})(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		results := make([]heartbeat.Result, len(hh))

		for i := range hh {
			results[i] = heartbeat.Result{Heartbeat: hh[i]}
		}

		return results, nil
	})

	results, err := handle(ctx, heartbeats)
	if err != nil {
		return nil, err
	}

	parsed := make([]heartbeat.Heartbeat, 0, len(results))
	for _, result := range results {
		parsed = append(parsed, result.Heartbeat)
	}

	return parsed, nil
}

func sendPreparedHeartbeats(
	ctx context.Context,
	v *viper.Viper,
	params params.Params,
	queueFilepath string,
	heartbeats []heartbeat.Heartbeat,
	withProjectConfig bool,
	paramsLoader func(context.Context, *viper.Viper, params.FlagReadOrder) (params.Params, error),
) error {
	logger := log.Extract(ctx)

	heartbeats = renderUserAgents(ctx, heartbeats)

	setLogFields(ctx, params)
	logger.Debugf("params: %s", params)

	if ratelimit.IsRateLimited(ratelimit.Params{
		Disabled:   params.Offline.Disabled,
		LastSentAt: params.Offline.LastSentAt,
		Timeout:    params.Offline.RateLimit,
	}) {
		err := offlinecmd.SaveHeartbeatsWithParams(ctx, v, queueFilepath, heartbeats, params)
		if err == nil {
			return nil
		}

		logger.Errorf("failed to save rate limited heartbeats: %s", err)
	}

	var (
		chOfflineSave = make(chan bool)
		savedOffline  bool
	)

	if len(heartbeats) > offline.SendLimit {
		savedOffline = true

		extraHeartbeats := heartbeats[offline.SendLimit:]

		logger.Debugf("save %d extra heartbeat(s) to offline queue", len(extraHeartbeats))

		go func(done chan<- bool) {
			if err := offlinecmd.SaveHeartbeatsWithParams(ctx, v, queueFilepath, extraHeartbeats, params); err != nil {
				logger.Errorf("failed to save extra heartbeats to offline queue: %s", err)
			}

			done <- true
		}(chOfflineSave)

		heartbeats = heartbeats[:offline.SendLimit]
	}

	sender, err := buildHandle(ctx, v, params, queueFilepath)
	if err != nil {
		if err := offlinecmd.SaveHeartbeatsWithParams(ctx, v, queueFilepath, heartbeats, params); err != nil {
			logger.Errorf("failed to save extra heartbeats to offline queue: %s", err)
		}

		if savedOffline {
			<-chOfflineSave
		}

		return fmt.Errorf("failed to initialize api client: %w", err)
	}

	opts := initHandleOptions()

	handle := sender
	if withProjectConfig {
		handle = handler.New(v, handler.Config{
			Params:       params,
			ParamsLoader: paramsLoader,
			Opts:         opts,
		})(sender)
	} else {
		for i := len(opts) - 1; i >= 0; i-- {
			handle = opts[i](params)(handle)
		}
	}

	results, err := handle(ctx, heartbeats)

	if savedOffline {
		<-chOfflineSave
	}

	if err != nil {
		return err
	}

	for _, result := range results {
		if len(result.Errors) > 0 {
			logger.Warnln(strings.Join(result.Errors, " "))
		}
	}

	if err := ratelimit.Reset(ctx, v); err != nil {
		logger.Errorf("failed to reset rate limit: %s", err)
	}

	return nil
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

func shouldUseProjectConfig(heartbeats []heartbeat.Heartbeat) bool {
	for _, h := range heartbeats {
		if h.EntityType == heartbeat.FileType && !h.IsUnsavedEntity {
			return true
		}
	}

	return false
}
