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
                func (c *Client) SummaryToday(userAgent string) ([]heartbeat.Summary, error)
        arguemnts/
            arguemnts.go
                type Args struct {}
        config/
            config.go
                type Config struct {}
        file/
            file.go
                type Stats struct {}
                func Info(filepath string) (Stats, error)
        heartbeat/
            heartbeat.go
                type Heartbeat struct {}
                type Summary struct {}
                type SendResult struct {}
                type Sender interface {
                    Send(hh []Heartbeat, hostName, userAgent string) ([]SendResult, error)
                }
            sanitize.go
                func Sanitize(h Heartbeat, obfuscate Obfuscate) Heartbeat
        log/
            log.go
                type LogLevel int32
                func Infof(msg string, args ...interface{})
                func Debugf(msg string, args ...interface{})
                func Fatalf(msg string, args ...interface{})
        offline/
            offline.go
                type Queue struct {}
                func (q *Queue) Send(hh []heartbeat.Heartbeat, hostName, userAgent string) ([]heartbeat.SendResult, error) 
                func WithQueue(s heartbeat.Sender, c Config) *Queue
        project/
            project.go
                func Info(filepath string) (project, branch string, err error)
                func InfoWithPlugin(filepath string, plugin Plugin) (project, branch string, err error)
        validate/
            validate.go
                func ByPattern(entity string, include, exclude []*regexp.Regexp) error
                func File(filepath string, includeOnlyWithProjectFile bool) error
    go.mod
    go.sum
    main.go
