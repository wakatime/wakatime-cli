package remote_test

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/remote"

	"github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestRegex(t *testing.T) {
	tests := map[string]struct {
		RemoteAddress string
		Expected      bool
	}{
		"ssh full path": {
			RemoteAddress: "ssh://user:1234@192.168.1.2/home/pi/unicorn-hat/examples/ascii_pic.py",
			Expected:      true,
		},
		"sftp full path": {
			RemoteAddress: "sftp://user:1234@192.168.1.2/home/pi/unicorn-hat/examples/ascii_pic.py",
			Expected:      true,
		},
		"without path": {
			RemoteAddress: "ssh://user:1234@192.168.1.2",
			Expected:      true,
		},
		"invalid ftp": {
			RemoteAddress: "ftp://user:1234@192.168.1.2",
			Expected:      false,
		},
		"invalid": {
			RemoteAddress: "http://192.168.1.2",
			Expected:      false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ok := remote.RemoteAddressRegex.MatchString(test.RemoteAddress)

			assert.Equal(t, test.Expected, ok)
		})
	}
}

func TestNewClient(t *testing.T) {
	client, err := remote.NewClient(
		"ssh://wakatime@192.168.1.2:222/home/pi/unicorn-hat/examples/ascii_pic.py",
		remote.Config{
			ConfigFile:   "~/.ssh/config",
			IdentityFile: "~/.ssh/id_rsa",
		})
	require.NoError(t, err)

	assert.Equal(t, remote.Client{
		ConfigFile:   "~/.ssh/config",
		Host:         "192.168.1.2",
		IdentityFile: "~/.ssh/id_rsa",
		Path:         "/home/pi/unicorn-hat/examples/ascii_pic.py",
		Port:         222,
		User:         "wakatime",
	}, client)
}

func TestNewClient_Sftp(t *testing.T) {
	client, err := remote.NewClient("sftp://127.0.0.1", remote.Config{})
	require.NoError(t, err)

	assert.Equal(t, remote.Client{
		ConfigFile:   "",
		Host:         "127.0.0.1",
		IdentityFile: "",
		Path:         "",
		Port:         22,
		User:         "",
	}, client)
}

func TestNewClient_Err(t *testing.T) {
	_, err := remote.NewClient("ssh://wakatime@192.168.1.2:port", remote.Config{})
	require.Error(t, err)

	assert.EqualError(t, err,
		`failed to parse remote file url: parse "ssh://wakatime@192.168.1.2:port": invalid port ":port" after host`)
}

func TestWithDetection(t *testing.T) {
	skipIfBinaryNotFoundOrWindows(t)

	host, port, shutdown := setupTestServer(t)
	defer shutdown()

	tests := map[string]struct {
		Hostname string
	}{
		"localhost": {
			Hostname: fmt.Sprintf("127.0.0.1:%d", port),
		},
		"domain name": {
			Hostname: "example.com",
		},
	}

	tmpDir := t.TempDir()

	knownHostFile := transformTemplateFile(t, "testdata/known_hosts_template", tmpDir, port)
	sshConfigFile := transformTemplateFile(t, "testdata/ssh_config_template", tmpDir, host, port, knownHostFile)

	entity, _ := filepath.Abs("testdata/main.go")

	opts := []heartbeat.HandleOption{
		remote.WithDetection(remote.Config{
			ConfigFile:   sshConfigFile,
			IdentityFile: "testdata/id_rsa",
		}),
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			sender := mockSender{
				SendHeartbeatsFn: func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
					assert.Equal(t, []heartbeat.Heartbeat{
						{
							Category:   heartbeat.CodingCategory,
							Entity:     fmt.Sprintf("ssh://%s%s", test.Hostname, entity),
							EntityRaw:  fmt.Sprintf("ssh://%s%s", test.Hostname, entity),
							EntityType: heartbeat.FileType,
							LocalFile:  hh[0].LocalFile,
							Time:       1585598060,
							UserAgent:  "wakatime/13.0.7",
						},
					}, hh)
					assert.Contains(t, hh[0].LocalFile, "main.go")
					return []heartbeat.Result{
						{
							Status:    201,
							Heartbeat: heartbeat.Heartbeat{},
						},
					}, nil
				},
			}

			handle := heartbeat.NewHandle(&sender, opts...)
			_, err := handle([]heartbeat.Heartbeat{
				{
					Category:   heartbeat.CodingCategory,
					Entity:     fmt.Sprintf("ssh://%s%s", test.Hostname, entity),
					EntityType: heartbeat.FileType,
					Time:       1585598060,
					UserAgent:  "wakatime/13.0.7",
				},
			})
			require.NoError(t, err)
		})
	}
}

type mockSender struct {
	SendHeartbeatsFn        func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error)
	SendHeartbeatsFnInvoked bool
}

func (m *mockSender) SendHeartbeats(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
	m.SendHeartbeatsFnInvoked = true
	return m.SendHeartbeatsFn(hh)
}

func setupTestServer(t *testing.T) (string, int, func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	host := ln.Addr().(*net.TCPAddr).IP.String()
	port := ln.Addr().(*net.TCPAddr).Port

	shutdown := make(chan struct{})

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-shutdown:
				default:
					t.Error("ssh server socket closed:", err)
				}

				return
			}

			go func() {
				defer conn.Close()

				sshServerConfig := newSshServerConfig(t)

				// from a standard TCP connection to an encrypted SSH connection
				sshconn, newChans, reqChans, err := ssh.NewServerConn(conn, sshServerConfig)
				if err != nil {
					t.Errorf("failed to handshake: %s", err)
					return
				}

				go ssh.DiscardRequests(reqChans)
				go handleNewChannels(newChans)

				_ = sshconn.Wait()
			}()
		}
	}()

	return host, port, func() {
		close(shutdown)
		ln.Close()
	}
}

func newSshServerConfig(t *testing.T) *ssh.ServerConfig {
	config := &ssh.ServerConfig{
		Config: ssh.Config{
			MACs: []string{"hmac-sha1"},
		},
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			return nil, nil
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			return &ssh.Permissions{
				Extensions: map[string]string{
					"pubkey-fp": ssh.FingerprintSHA256(key),
				},
			}, nil
		},
	}

	privKey := []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEArhp7SqFnXVZAgWREL9Ogs+miy4IU/m0vmdkoK6M97G9NX/Pj
wf8I/3/ynxmcArbt8Rc4JgkjT2uxx/NqR0yN42N1PjO5Czu0dms1PSqcKIJdeUBV
7gdrKSm9Co4d2vwfQp5mg47eG4w63pz7Drk9+VIyi9YiYH4bve7WnGDswn4ycvYZ
slV5kKnjlfCdPig+g5P7yQYud0cDWVwyA0+kxvL6H3Ip+Fu8rLDZn4/P1WlFAIuc
PAf4uEKDGGmC2URowi5eesYR7f6GN/HnBs2776laNlAVXZUmYTUfOGagwLsEkx8x
XdNqntfbs2MOOoK+myJrNtcB9pCrM0H6um19uQIDAQABAoIBABkWr9WdVKvalgkP
TdQmhu3mKRNyd1wCl+1voZ5IM9Ayac/98UAvZDiNU4Uhx52MhtVLJ0gz4Oa8+i16
IkKMAZZW6ro/8dZwkBzQbieWUFJ2Fso2PyvB3etcnGU8/Yhk9IxBDzy+BbuqhYE2
1ebVQtz+v1HvVZzaD11bYYm/Xd7Y28QREVfFen30Q/v3dv7dOteDE/RgDS8Czz7w
jMW32Q8JL5grz7zPkMK39BLXsTcSYcaasT2ParROhGJZDmbgd3l33zKCVc1zcj9B
SA47QljGd09Tys958WWHgtj2o7bp9v1Ufs4LnyKgzrB80WX1ovaSQKvd5THTLchO
kLIhUAECgYEA2doGXy9wMBmTn/hjiVvggR1aKiBwUpnB87Hn5xCMgoECVhFZlT6l
WmZe7R2klbtG1aYlw+y+uzHhoVDAJW9AUSV8qoDUwbRXvBVlp+In5wIqJ+VjfivK
zgIfzomL5NvDz37cvPmzqIeySTowEfbQyq7CUQSoDtE9H97E2wWZhDkCgYEAzJdJ
k+NSFoTkHhfD3L0xCDHpRV3gvaOeew8524fVtVUq53X8m91ng4AX1r74dCUYwwiF
gqTtSSJfx2iH1xKnNq28M9uKg7wOrCKrRqNPnYUO3LehZEC7rwUr26z4iJDHjjoB
uBcS7nw0LJ+0Zeg1IF+aIdZGV3MrAKnrzWPixYECgYBsffX6ZWebrMEmQ89eUtFF
u9ZxcGI/4K8ErC7vlgBD5ffB4TYZ627xzFWuBLs4jmHCeNIJ9tct5rOVYN+wRO1k
/CRPzYUnSqb+1jEgILL6istvvv+DkE+ZtNkeRMXUndWwel94BWsBnUKe0UmrSJ3G
sq23J3iCmJW2T3z+DpXbkQKBgQCK+LUVDNPE0i42NsRnm+fDfkvLP7Kafpr3Umdl
tMY474o+QYn+wg0/aPJIf9463rwMNyyhirBX/k57IIktUdFdtfPicd2MEGETElWv
nN1GzYxD50Rs2f/jKisZhEwqT9YNyV9DkgDdGGdEbJNYqbv0qpwDIg8T9foe8E1p
bdErgQKBgAt290I3L316cdxIQTkJh1DlScN/unFffITwu127WMr28Jt3mq3cZpuM
Aecey/eEKCj+Rlas5NDYKsB18QIuAw+qqWyq0LAKLiAvP1965Rkc4PLScl3MgJtO
QYa37FK0p8NcDeUuF86zXBVutwS5nJLchHhKfd590ks57OROtm29
-----END RSA PRIVATE KEY-----
	`)

	signer, err := ssh.ParsePrivateKey(privKey)
	if err != nil {
		t.Fatalf("Failed to parse private key: %s", err)
	}

	config.AddHostKey(signer)

	return config
}

func handleNewChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go handleNewChannel(newChannel)
	}
}

func handleNewChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		_ = newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	ch, reqs, err := newChannel.Accept()
	if err != nil {
		fmt.Printf("could not accept channel: %s", err)
		return
	}

	go func() {
		for req := range reqs {
			switch req.Type {
			case "subsystem":
				err = handleSubsystem(ch, req)
				if err != nil {
					fmt.Printf("could not handle subsystem: %s", err)
				}
			default:
				fmt.Printf("unknown request type: %s", req.Type)

				_ = req.Reply(false, nil)
			}
		}
	}()
}

func handleSubsystem(ch ssh.Channel, req *ssh.Request) error {
	defer func() {
		_ = ch.CloseWrite()
		_ = ch.Close()
	}()

	var sshreq struct {
		Name string
	}

	if err := ssh.Unmarshal(req.Payload, &sshreq); err != nil {
		return req.Reply(false, nil)
	}

	if sshreq.Name != "sftp" {
		return req.Reply(false, nil)
	}

	_ = req.Reply(true, nil)

	sftpServer, err := sftp.NewServer(ch)
	if err != nil {
		return err
	}

	// wait for the session to close
	err = sftpServer.Serve()

	exitStatus := uint32(1)
	if err == nil {
		exitStatus = uint32(0)
	}

	var sshRes = struct {
		Status uint32
	}{
		Status: exitStatus,
	}

	_, exitStatusErr := ch.SendRequest("exit-status", false, ssh.Marshal(sshRes))

	return exitStatusErr
}

func transformTemplateFile(t *testing.T, fp, tmpDir string, args ...interface{}) string {
	tplFile, err := os.ReadFile(fp)
	require.NoError(t, err)

	f, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer f.Close()

	tplStr := fmt.Sprintf(string(tplFile), args...)

	_, err = f.WriteString(tplStr)
	require.NoError(t, err)

	return f.Name()
}

func findSftpBinary() bool {
	// sftp is available in unix-like systems and windows 10+
	_, err := exec.LookPath("sftp")
	if err != nil {
		return false
	}

	return true
}

func skipIfBinaryNotFoundOrWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping because OS is not windows.")
	}

	if !findSftpBinary() {
		t.Skip("Skipping because sftp binary is not installed in this machine.")
	}
}
