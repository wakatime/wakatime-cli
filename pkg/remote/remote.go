package remote

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// RemoteAddressRegex is a pattern for (ssh|sftp)://user:pass@host:port.
var RemoteAddressRegex = regexp.MustCompile(`(?i)^((ssh|sftp)://)+(?P<credentials>[^:@]+(:([^:@])+)?@)?[^:]+(:\d+)?`)

const (
	defaultPort        = 22
	defaultTimeoutSecs = 20
	// Max file size supporting downloading from remote. Default is 512Kb.
	maxFileSize = 512000
)

// Client communicates using sftp protocol.
type Client struct {
	User string
	Pass string
	Host string
	Port int
	Path string
}

// WithDetection initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to detect remote file and
// download to a temporary directory.
func WithDetection() heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			log.Debugln("execute remote file detection")

			var (
				tmpDir string
				err    error
			)

			for n, h := range hh {
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

						return next(hh)
					}
				}

				tmpFile, err := os.CreateTemp(tmpDir, fmt.Sprintf("*%s", filepath.Base(h.Entity)))
				if err != nil {
					log.Errorf("failed to create temporary file: %s", err)

					continue
				}

				c, err := NewClient(h.Entity)
				if err != nil {
					log.Errorf("failed to create new remote client: %s", err)

					continue
				}

				err = c.DownloadFile(tmpFile.Name())
				if err != nil {
					log.Errorf("failed to download file to temporary folder: %s", err)

					continue
				}

				hh[n].LocalFile = tmpFile.Name()
				// we save untouched entity for offline handling
				hh[n].EntityRaw = h.Entity
			}

			return next(hh)
		}
	}
}

// NewClient initializes a new remote client.
func NewClient(address string) (Client, error) {
	parsedURL, err := url.Parse(address)
	if err != nil {
		return Client{}, fmt.Errorf("failed to parse remote file url: %s", err)
	}

	host := parsedURL.Host
	pass, _ := parsedURL.User.Password()
	port := defaultPort

	if parsedURL.Port() != "" {
		// we're safe to ignore error here since `url.Parse` checks if port is valid integer
		port, _ = strconv.Atoi(parsedURL.Port())
		host = strings.Split(host, ":")[0]
	}

	return Client{
		User: parsedURL.User.Username(),
		Pass: pass,
		Host: host,
		Port: port,
		Path: parsedURL.Path,
	}, nil
}

// DownloadFile downloads a remote file and copy to a local file.
func (c Client) DownloadFile(localFile string) error {
	conn, sc, err := c.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to sftp host: %s", err)
	}

	defer conn.Close()
	defer sc.Close()

	srcFile, err := sc.OpenFile(c.Path, os.O_RDONLY)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %s", err)
	}

	defer srcFile.Close()

	dstFile, err := os.Create(localFile)
	if err != nil {
		return fmt.Errorf("failed to open local file: %s", err)
	}

	defer dstFile.Close()

	_, err = io.CopyN(dstFile, srcFile, maxFileSize)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to download remote file: %s", err)
	}

	return nil
}

// Connect connects to sftp host.
func (c Client) Connect() (*ssh.Client, *sftp.Client, error) {
	hostKeys, err := getHostKeys(c.Host)
	if err != nil {
		log.Errorf("failed to get host keys: %s", err)
	}

	var auths []ssh.AuthMethod

	// Try to use $SSH_AUTH_SOCK which contains the path of the unix file socket that the sshd agent uses
	// for communication with other processes
	if aconn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(aconn).Signers))
	}

	// Use password authentication if provided
	if c.Pass != "" {
		auths = append(auths, ssh.Password(c.Pass))
	}

	// Initialize client configuration
	config := ssh.ClientConfig{
		User:    c.User,
		Auth:    auths,
		Timeout: defaultTimeoutSecs * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)

	var conn *ssh.Client

	if len(hostKeys) == 0 {
		log.Warnf("no host key found for %s. It will try to make an insecure connection", c.Host)

		config.HostKeyCallback = ssh.InsecureIgnoreHostKey() // nolint:gosec

		// Connect to server
		conn, err = dial(addr, &config)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to connect to '%s': %s", addr, err)
		}
	} else {
		log.Debugf("found %d ssh keys for host. It will loop over them and try to connect", len(hostKeys))

		for _, hostKey := range hostKeys {
			config.HostKeyCallback = ssh.FixedHostKey(hostKey)

			// Connect to server
			conn, err = dial(addr, &config)
			if err != nil {
				log.Warnf("failed to connect to '%s': %s", addr, err)

				continue
			}

			break
		}
	}

	// Create new SFTP client
	sc, err := sftp.NewClient(conn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start sftp subsystem: %s", err)
	}

	return conn, sc, nil
}

// getHostKeys gets all host keys from local known hosts for given host.
func getHostKeys(host string) ([]ssh.PublicKey, error) {
	// parse OpenSSH known_hosts file ssh or use ssh-keyscan to get initial key
	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		return nil, fmt.Errorf("failed to read known_hosts files: %s", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	hostKeys := []ssh.PublicKey{}

	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}

		if strings.Contains(fields[0], host) {
			var err error

			hostKey, _, _, _, err := ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				log.Warnf("failed to parse %q: %s", fields[2], err)
				continue
			}

			hostKeys = append(hostKeys, hostKey)
		}
	}

	return hostKeys, nil
}

func dial(addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial to '%s': %s", addr, err)
	}

	return conn, nil
}
