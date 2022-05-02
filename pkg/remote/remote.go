package remote

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// RemoteAddressRegex is a pattern for (ssh|sftp)://user:pass@host:port.
var RemoteAddressRegex = regexp.MustCompile(`(?i)^((ssh|sftp)://)+(?P<credentials>[^:@]+(:([^:@])+)?@)?[^:]+(:\d+)?`)

const (
	defaultPort        = 22
	defaultTimeoutSecs = "20"
)

// Client communicates using sftp protocol.
type (
	Config struct {
		ConfigFile   string
		IdentityFile string
	}

	Client struct {
		ConfigFile   string
		Host         string
		IdentityFile string
		Path         string
		Port         int
		User         string
	}
)

// WithDetection initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to detect remote file and
// download to a temporary directory.
func WithDetection(config Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			log.Debugln("execute remote file detection")

			var (
				tmpDir string
				err    error
			)

			for i, h := range hh {
				if h.EntityType != heartbeat.FileType {
					continue
				}

				if h.IsUnsavedEntity {
					continue
				}

				if !RemoteAddressRegex.MatchString(h.Entity) {
					continue
				}

				if tmpDir == "" {
					tmpDir, err = os.MkdirTemp(os.TempDir(), "")
					if err != nil {
						log.Errorf("failed to create temporary directory: %s", err)

						continue
					}

					defer os.RemoveAll(tmpDir)
				}

				tmpFile, err := os.CreateTemp(tmpDir, fmt.Sprintf("*%s", filepath.Base(h.Entity)))
				if err != nil {
					log.Errorf("failed to create temporary file: %s", err)

					continue
				}

				c, err := NewClient(h.Entity, config)
				if err != nil {
					log.Errorf("failed to create new remote client: %s", err)

					continue
				}

				err = c.DownloadFile(tmpFile.Name())
				if err != nil {
					log.Errorf("failed to download file to temporary folder: %s", err)

					continue
				}

				hh[i].LocalFile = tmpFile.Name()
				// we save untouched entity for offline handling
				hh[i].EntityRaw = h.Entity
			}

			return next(hh)
		}
	}
}

// NewClient initializes a new remote client.
func NewClient(address string, config Config) (Client, error) {
	parsedURL, err := url.Parse(address)
	if err != nil {
		return Client{}, fmt.Errorf("failed to parse remote file url: %s", err)
	}

	host := parsedURL.Host
	port := defaultPort

	if parsedURL.Port() != "" {
		// we're safe to ignore error here since `url.Parse` checks if port is valid integer
		port, _ = strconv.Atoi(parsedURL.Port())
		host = strings.Split(host, ":")[0]
	}

	host = strings.TrimSuffix(host, ":")

	return Client{
		ConfigFile:   config.ConfigFile,
		Host:         host,
		IdentityFile: config.IdentityFile,
		Path:         parsedURL.Path,
		Port:         port,
		User:         parsedURL.User.Username(),
	}, nil
}

// DownloadFile downloads a remote file and copy to a local file.
func (c Client) DownloadFile(localFile string) error {
	sftpbin, err := findSftpBinary()
	if err != nil {
		return fmt.Errorf("failed to find scp binary: %s", err)
	}

	// -C - enables compression
	args := []string{"-C", "-o", "ConnectTimeout=" + defaultTimeoutSecs}

	if c.ConfigFile != "" {
		args = append(args, "-F", c.ConfigFile)
	}

	if c.IdentityFile != "" {
		args = append(args, "-i", c.IdentityFile)
	}

	if c.Port != defaultPort {
		args = append(args, "-P", strconv.Itoa(c.Port))
	}

	if c.User != "" {
		args = append(args, fmt.Sprintf("%s@%s:%s", c.User, c.Host, c.Path))
	} else {
		args = append(args, fmt.Sprintf("%s:%s", c.Host, c.Path))
	}

	args = append(args, localFile)

	log.Debugf("ssh command args: %s", strings.Join(args, " "))

	cmd := exec.Command(sftpbin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to download file: %s", err)
	}

	return nil
}

func findSftpBinary() (string, error) {
	bin, err := exec.LookPath("sftp")
	if err != nil {
		return "", fmt.Errorf("sftp binary not found: %s", err)
	}

	return bin, nil
}
