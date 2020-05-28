wakatime-cli/
    cmd/
        legacy/
            configread/
                configread.go
            configwrite/
                configwrite.go
            heartbeat/
                heartbeat.go
            today/
                today.go
            version.go
            run.go
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
                func WithUserAgent(plugin string) (Option, error)
                func WithUserAgentUnknownPlugin() (Option, error)

                type Client struct {}
                func NewClient(baseURL string, client *http.Client, opts ...Option)
                func (c *Client) Send(hh []heartbeat.Heartbeat) ([]heartbeat.SendResult, error)
                func (c *Client) Summaries(startDate, endDate time.Time) ([]summary.Summary, error)
        deps/
            deps.go
                WithDetection() heartbeat.HandleOption
                WithDetectionOnFile(filepath string) heartbeat.HandleOption
                func detect(filepath, language string) (deps []string, err error)
        filestats/
            filestats.go
                WithDetection() heartbeat.HandleOption
                WithDetectionOnFile(filepath string) heartbeat.HandleOption
                func detect(filepath string) (lines int, err error)
        heartbeat/
            heartbeat.go
                type Heartbeat struct {}
                type Result struct {}
                type Sender interface {
                    Send(hh []Heartbeat) ([]Result, error)
                }
                type Handle func(hh []Heartbeat) ([]Result, error)
                HandleOption func(Handle) Handle
                func NewHandle(Sender, ...HandleOption) Handle
            sanitize.go
                type SanitizeConfig struct {
                    HideBranchNames []*regexp.Regexp
                    HideFileNames []*regexp.Regexp
                    HideProjectNames []*regexp.Regexp
                }
                WithSanitization(c SanitizeConfig) heartbeat.HandleOption
                func sanitize(h Heartbeat, obfuscate Obfuscate) Heartbeat
            validate.go
                type ValidateConfig struct {
                    Exclude []*regexp.Regexp
                    ExcludeUnknownProject bool
                    Include []*regexp.Regexp
                    IncludeOnlyWithProjectFile bool
                }
                WithValidation(c ValidateConfig) heartbeat.HandleOption
                func validateByPattern(entity string, include, exclude []*regexp.Regexp) error
                func validateFile(filepath string, includeOnlyWithProjectFile bool) error
        language/
            language.go
                type Config struct {
                    Alternative string
                    Overwrite string
                    LocalFile string
                }
                WithDetection(c Config) heartbeat.HandleOption
                func detect(filepath string) (language string, err error)
        offline/
            offline.go
                func WithQueue(filepath, table string) heartbeat.HandleOption
        project/
            project.go
                type Config struct {
                    Alternative string
                    Overwrite string
                    LocalFile string
                }
                WithDetection(c Config) heartbeat.HandleOption
                func detect(filepath string) (project, branch string, err error)
                func detectWithPlugin(filepath string, plugin Plugin) (project, branch string, err error)
        summary/
            summary.go
                type Summary struct {}
        version/
            version.go
    go.mod
    go.sum
    main.go
