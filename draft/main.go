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
)

const (
	queueDBFile = ".wakatime.db"
	queueDBTable = "heartbeat_2"
)

func main() {
	withAuth, err := api.WithAuth(api.BasicAuth{
		Secret: args.APIKey,
	})
	if err != nil {
		log.Fatalf(err)
	}

	clientOpts := []api.Option{
		withAuth,
		api.WithHostName(args.HostName),
	}

	if args.SSLCert != nil {
		opts = append(options, api.WithSSL(args.SSLCert))
	}

	if args.Timeout != nil {
		opts = append(options, api.WithTimeout(args.Timeout * time.Second))
	}

	if args.Plugin != nil {
		opts = append(options, api.WithUserAgentFromPlugin(args.Plugin))
	} else {
		opts = append(options, api.WithUserAgent())
	}

	client = api.NewClient(baseURL, http.DefaultClient, clientOpts...)

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
	sender := NewSender(client, senderOpts...)

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
	_, err := sender.Send(hh)
	if err != nil {
		log.Fatalf(err)
	}
}
