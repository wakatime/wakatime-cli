import (
f	"net/http"
	"time"

	"github.com/alanhamlett/wakatime-cli/pkg/api"
	"github.com/alanhamlett/wakatime-cli/pkg/file"
	"github.com/alanhamlett/wakatime-cli/pkg/log"
	"github.com/alanhamlett/wakatime-cli/pkg/heartbeat"
	"github.com/alanhamlett/wakatime-cli/pkg/project"
	"github.com/alanhamlett/wakatime-cli/pkg/offline"
	"github.com/alanhamlett/wakatime-cli/pkg/validate"
)

func main() {
	// initialize api client
	opts := []api.Option{
		api.Auth("", args.APIKey)
	}

	if args.Timeout != nil {
		opts = append(options, api.Timeout(args.Timeout * time.Second))
	}

	if args.SSLCert != nil {
		opts = append(options, api.SSL(args.SSLCert))
	}

	var client heartbeat.Sender
	client = api.NewClient(baseURL, http.DefaultClient, ...opts)
	client = offline.WithQueue(client, offline.Config{})

	// detect project info and file stats
	var (
		branch string
		project string
		stats file.Stats
		err error
	)
	if args.EntityType = "file" {
		branch, project, err = project.Info(args.Entity, args.AlternativeProjectName)
		if err != nil {
			log.Fatalf(err)
		}
		stats, err = file.Info(args.Entity)
		if err != nil {
			log.Fatalf(err)
		}
	}

	// validate data
	if args.ExcludeUnknownProject && args.Project == "" {
		log.Debugf("skipping heartbeat: project file missing")
		return
	}

	if err := validate.ByPattern(args.Entity, args.Include, args.Exclude); err != nil {
		log.Debugf("skipping heartbeat: %s", err)
	}

	if args.EntityType == "file" {
		if err := validate.File(args.Entity, args.IncludeOnlyWithProjectFile); err != nil {
			log.Debugf("skipping heartbeat: %s", err)
			return
		}
	}

	// send heartbeat
	hh := []Heartbeat{
		{
			Branch:         branch,
			Category:       args.Category,
			CursorPosition: stats.CursorPosition,
			Dependencies:   stats.Dependencies,
			Entity:         args.Entity,
			EntityType:     args.EntityType,
			IsWrite:        args.IsWrite,
			Language:       stats.Language,
			LineNumber:     stats.LineNumber,
			Lines:          stats.Lines,
			Project:        project,
			Time:           args.Time,
			UserAgent:      arg.UserAgent,
		}
	}	
	_, err := client.Send(hh, heartbeat.SendConfig{})
	if err != nil {
		log.Fatalf(err)
	}
}
