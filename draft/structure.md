wakatime-cli/
    cmd/
        root.go
    pkg/
        api/
            api.go
                type Option func(*Client)

                func Auth(user, secret string) Option
                func NTLM(proxy string) Option
                func SSL(cert string) Option
                func Timeout(timeout time.Duration) Option

                type Client struct {}
                func NewClient(baseURL string, client *http.Client, opts ...Option)
                func (c *Client) Send(hh []heartbeat.Heartbeat, hostName, userAgent string) ([]heartbeat.SendResult, error)
                func (c *Client) Summaries(start, end time.Time, userAgent string) ([]heartbeat.Summary, error)
        arguments/
            arguments.go
                type Args struct {}
        config/
            config.go
                type Config struct {}
        deps/
            deps.go
                type Config stuct {
                    LocalFile string
                }
                Detect(c Config) heartbeat.SenderOption
                func detect(filepath, language string) (deps []string, err error)
                type sender struct {}
                func (d *sender) Send(hh []heartbeat.Heartbeat, hostName, userAgent string) ([]heartbeat.SendResult, error)
        filestats/
            filestats.go
                type Config stuct {
                    LocalFile string
                }
                Detect(c Config) heartbeat.SenderOption
                func detect(filepath string) (lines int, err error)
                type sender struct {}
                func (d *sender) Send(hh []heartbeat.Heartbeat, hostName, userAgent string) ([]heartbeat.SendResult, error)
        heartbeat/
            heartbeat.go
                type Heartbeat struct {}
                type SendResult struct {}
                type Summary struct {}
                type Sender interface {
                    Send(hh []Heartbeat, hostName, userAgent string) ([]SendResult, error)
                }
                SenderOption func(Sender) Sender
                func NewSender(sender, ...[]SenderOption) Sender
            sanitize.go
                type SanitizeConfig stuct {
                    HideBranchNames []*regexp.Regexp
                    HideFileNames []*regexp.Regexp
                    HideProjectNames []*regexp.Regexp
                }
                Sanitize(c SanitizeConfig) heartbeat.SenderOption
                func sanitize(h Heartbeat, obfuscate Obfuscate) Heartbeat
                type sanitizeSender struct {}
                func (s *sanitizeSender) Send(hh []heartbeat.Heartbeat, hostName, userAgent string) ([]heartbeat.SendResult, error)
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
                func (v *validateSender) Send(hh []heartbeat.Heartbeat, hostName, userAgent string) ([]heartbeat.SendResult, error)
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
                func (d *sender) Send(hh []heartbeat.Heartbeat, hostName, userAgent string) ([]heartbeat.SendResult, error)
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
                func (q *sender) Send(hh []heartbeat.Heartbeat, hostName, userAgent string) ([]heartbeat.SendResult, error)
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
                func (s *sender) Send(hh []heartbeat.Heartbeat, hostName, userAgent string) ([]heartbeat.SendResult, error)
    go.mod
    go.sum
    main.go
