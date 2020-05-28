import (
	"net/http"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/filestats"
	"github.com/wakatime/wakatime-cli/pkg/deps"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/language"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/project"
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
		clientOpts = append(clientOpts, api.WithSSL(args.SSLCert))
	}

	if args.Timeout != nil {
		clientOpts = append(clientOpts, api.WithTimeout(args.Timeout * time.Second))
	}

	if args.Plugin != nil {
		clientOpts = append(clientOpts, api.WithUserAgentFromPlugin(args.Plugin))
	} else {
		clientOpts = append(clientOpts, api.WithUserAgent())
	}

	client := api.NewClient(baseURL, http.DefaultClient, clientOpts...)

	var withDepsDetection heartbeat.HandleOption
	if args.Localfile == "" {
		withDepsDetection = deps.WithDetection()
	} else
		withDepsDetection = deps.WithDetectionOnFile(args.Localfile)
	}

	var withFilestatsDetection heartbeat.HandleOption
	if args.Localfile == "" {
		withFilestatsDetection = filestats.WithDetection()
	} else
		withFilestatsDetection = filestats.WithDetectionOnFile(args.Localfile)
	}

	handleOpts := []heartbeat.HandleOption{
		heartbeat.WithSanitization(heartbeat.SanitizeConfig{
			HideBranchNames: args.HideBranchNames,
			HideFileNames: args.HideFileNames,
			HideProjectNames: args.HideProjectNames,
		}),
		offline.WithQueue(queueDBFile, queueDBTable),
		language.WithDetection(language.Config{
			Alternative: args.AlternativeLanguage,
			Overwrite: args.Language,
			LocalFile: args.LocalFile,
		}),
		withDepsDetection,
		withFilestatsDetection,
		project.WithDetection(language.Config{
			Alternative: args.AlternativeProject,
			Overwrite: args.Project,
			LocalFile: args.LocalFile,
		}),
		heartbeat.WithValidation(heartbeat.ValidateConfig{
			Exclude: args.Exclude,
			ExcludeUnknownProject: args.ExcludeUnknownProject,
			Include: args.Include,
			IncludeOnlyWithProjectFile: args.IncludeOnlyWithProjectFile,
		),
	}
	handle := heartbeat.NewHandle(client, handleOpts...)

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
	_, err := handle(hh)
	if err != nil {
		log.Fatalf(err)
	}
}
