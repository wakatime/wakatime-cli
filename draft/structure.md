wakatime-cli/
    cmd/
        legacy/
            config.go
            heartbeat.go
            legacy.go
            today.go
            version.go
        root.go
    pkg/
        api/
            api.go
                type BasicAuth
                func (a BasicAuth) HeaderValue() (string, error)

                type Option func(*Client)

                func WithAuth(auth BasicAuth) (Option, error)
                func WithHostName(hostName string) Option
                func WithNTLM(proxy string) Option
                func WithSSL(cert string) Option
                func WithTimeout(timeout time.Duration) Option
                func WithUserAgent() (Option, error)
                func WithUserAgentFromPlugin(plugin string) (Option, error)

                type Client struct {}
                func NewClient(baseURL string, client *http.Client, opts ...Option)
                func (c *Client) Send(hh []heartbeat.Heartbeat) ([]heartbeat.SendResult, error)
                func (c *Client) Summaries(startDate, endDate time.Time) ([]summary.Summary, error)
        deps/
            deps.go
                type Config stuct {
                    LocalFile string
                }
                Detect(c Config) heartbeat.SenderOption
                func detect(filepath, language string) (deps []string, err error)
                type sender struct {}
                func (s *sender) Send(hh []heartbeat.Heartbeat) ([]heartbeat.SendResult, error)
        filestats/
            filestats.go
                type Config stuct {
                    LocalFile string
                }
                Detect(c Config) heartbeat.SenderOption
                func detect(filepath string) (lines int, err error)
                type sender struct {}
                func (s *sender) Send(hh []heartbeat.Heartbeat) ([]heartbeat.SendResult, error)
        heartbeat/
            heartbeat.go
                type Heartbeat struct {}
                type SendResult struct {}
                type Sender interface {
                    Send(hh []Heartbeat) ([]SendResult, error)
                }
                SenderOption func(Sender) Sender
                func NewSender(sender, ...SenderOption) Sender
            sanitize.go
                type SanitizeConfig stuct {
                    HideBranchNames []*regexp.Regexp
                    HideFileNames []*regexp.Regexp
                    HideProjectNames []*regexp.Regexp
                }
                Sanitize(c SanitizeConfig) heartbeat.SenderOption
                func sanitize(h Heartbeat, obfuscate Obfuscate) Heartbeat
                type sanitizeSender struct {}
                func (s *sanitizeSender) Send(hh []heartbeat.Heartbeat) ([]heartbeat.SendResult, error)
            validate.go
                type ValidateConfig stuct {
                    Exclude []*regexp.Regexp
                    ExcludeUnknownProject bool
                    Include []*regexp.Regexp
                    IncludeOnlyWithProjectFile bool
                }
                Validate(c ValidateConfig) heartbeat.SenderOption
                func validateByPattern(entity string, include, exclude []*regexp.Regexp) error
                func validateFile(filepath string, includeOnlyWithProjectFile bool) error
                type validateSender struct {}
                func (s *validateSender) Send(hh []heartbeat.Heartbeat) ([]heartbeat.SendResult, error)
        language/
            language.go
                type Config stuct {
                    Alternative string
                    Overwrite string
                    LocalFile string
                }
                Detect(c Config) heartbeat.SenderOption
                func detect(filepath string) (language string, err error)
                type sender struct {}
                func (s *sender) Send(hh []heartbeat.Heartbeat) ([]heartbeat.SendResult, error)
        log/
            log.go
                type LogLevel int32
                func Infof(msg string, args ...interface{})
                func Debugf(msg string, args ...interface{})
                func Fatalf(msg string, args ...interface{})
        offline/
            offline.go
                func Queue(filepath, table string) heartbeat.SenderOption
                type sender struct {}
                func (s *sender) Send(hh []heartbeat.Heartbeat) ([]heartbeat.SendResult, error)
        project/
            project.go
                type Config stuct {
                    Alternative string
                    Overwrite string
                    LocalFile string
                }
                Detect(c Config) heartbeat.SenderOption
                func detect(filepath string) (project, branch string, err error)
                func detectWithPlugin(filepath string, plugin Plugin) (project, branch string, err error)
                type sender struct {}
                func (s *sender) Send(hh []heartbeat.Heartbeat) ([]heartbeat.SendResult, error)
        summary/
            summary.go
                type Summary struct {}
        version/
            version.go
    go.mod
    go.sum
    main.go
