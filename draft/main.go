import (
	"net/http"
	"time"

	"github.com/alanhamlett/wakatime-cli/pkg/api"
	"github.com/alanhamlett/wakatime-cli/pkg/filestats"
	"github.com/alanhamlett/wakatime-cli/pkg/deps"
	"github.com/alanhamlett/wakatime-cli/pkg/heartbeat"
	"github.com/alanhamlett/wakatime-cli/pkg/language"
	"github.com/alanhamlett/wakatime-cli/pkg/log"
	"github.com/alanhamlett/wakatime-cli/pkg/offline"
	"github.com/alanhamlett/wakatime-cli/pkg/project"
	"github.com/alanhamlett/wakatime-cli/pkg/sanitize"
	"github.com/alanhamlett/wakatime-cli/pkg/validate"
)

const (
	queueDBFile = ".wakatime.db"
	queueDBTable = "heartbeat_2"
)

func main() {
	clientOpts := []api.Option{
		api.Auth("", args.APIKey)
	}

	if args.Timeout != nil {
		opts = append(options, api.Timeout(args.Timeout * time.Second))
	}

	if args.SSLCert != nil {
		opts = append(options, api.SSL(args.SSLCert))
	}

	client = api.NewClient(baseURL, http.DefaultClient, ...opts)

	senderOpts := []heartbeat.SenderOption{
		heartbeat.Sanitize(heartbeat.SanitizeConfig{
			HideBranchNames: args.HideBranchNames,
			HideFileNames: args.HideFileNames,
			HideProjectNames: args.HideProjectNames,
		}),
		offline.Queue(queueDBFile, queueDBTable),
		language.Detect(language.Config{
			Alternative: args.AlternativeLanguage,
			Overwrite: args.Language,
			LocalFile: args.LocalFile,
		}),
		deps.Detect(dep.Config{
			LocalFile: args.Localfile,
		}),
		filestats.Detect(filestats.Config{
			LocalFile: args.Localfile,
		}),
		project.Detect(language.Config{
			Alternative: args.AlternativeProject,
			Overwrite: args.Project,
			LocalFile: args.LocalFile,
		}),
		heartbeat.Validate(heartbeat.ValidateConfig{
			Exclude: args.Exclude,
			ExcludeUnknownProject: args.ExcludeUnknownProject,
			Include: args.Include,
			IncludeOnlyWithProjectFile: args.IncludeOnlyWithProjectFile,
		),
	}
	sender := NewSender(client, senderOpts)

	hh := []Heartbeat{
		{
			Category:       args.Category,
			Entity:         args.Entity,
			EntityType:     args.EntityType,
			IsWrite:        args.IsWrite,
			Time:           args.Time,
			UserAgent:      arg.UserAgent,
		}
	}
	_, err := sender.Send(hh, heartbeat.SendConfig{})
	if err != nil {
		log.Fatalf(err)
	}
}
