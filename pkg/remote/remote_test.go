package remote_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/filter"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/regex"
	"github.com/wakatime/wakatime-cli/pkg/remote"

	"github.com/kevinburke/ssh_config"
	"github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestNewClient(t *testing.T) {
	client, err := remote.NewClient("ssh://wakatime:1234@192.168.1.2:222/home/pi/unicorn-hat/examples/ascii_pic.py")
	require.NoError(t, err)

	assert.Equal(t, remote.Client{
		User:         "wakatime",
		Pass:         "1234",
		OriginalHost: "192.168.1.2",
		Host:         "192.168.1.2",
		Port:         222,
		Path:         "/home/pi/unicorn-hat/examples/ascii_pic.py",
	}, client)
}

func TestNewClient_Sftp(t *testing.T) {
	client, err := remote.NewClient("sftp://127.0.0.1")
	require.NoError(t, err)

	assert.Equal(t, remote.Client{
		User:         "",
		Pass:         "",
		OriginalHost: "127.0.0.1",
		Host:         "127.0.0.1",
		Port:         22,
		Path:         "",
	}, client)
}

func TestNewClient_Err(t *testing.T) {
	_, err := remote.NewClient("ssh://wakatime:1234@192.168.1.2:port")
	require.Error(t, err)

	assert.EqualError(t, err,
		`failed to parse remote file url: parse "ssh://wakatime:1234@192.168.1.2:port": invalid port ":port" after host`)
}

func TestWithDetection_SshConfig_Hostname(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping because OS is Windows.")
	}

	shutdown, host, port := testServer(t, false)
	defer shutdown()

	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	ssh_config.DefaultUserSettings = &ssh_config.UserSettings{
		IgnoreErrors: false,
	}

	ssh_config.DefaultUserSettings.ConfigFinder(func() string {
		return tmpFile.Name()
	})

	template, err := os.ReadFile("testdata/ssh_config_hostname")
	require.NoError(t, err)

	err = os.WriteFile(tmpFile.Name(), []byte(fmt.Sprintf(string(template), host)), 0600)
	require.NoError(t, err)

	entity, _ := filepath.Abs("./testdata/main.go")

	sender := mockSender{
		SendHeartbeatsFn: func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			assert.Equal(t, []heartbeat.Heartbeat{
				{
					Category:              heartbeat.CodingCategory,
					Entity:                "ssh://user:pass@example.com:" + strconv.Itoa(port) + entity,
					EntityType:            heartbeat.FileType,
					LocalFile:             hh[0].LocalFile,
					LocalFileNeedsCleanup: true,
					Time:                  1585598060,
					UserAgent:             "wakatime/13.0.7",
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

	opts := []heartbeat.HandleOption{
		remote.WithDetection(),
	}

	handle := heartbeat.NewHandle(&sender, opts...)
	_, err = handle([]heartbeat.Heartbeat{
		{
			Category:   heartbeat.CodingCategory,
			Entity:     "ssh://user:pass@example.com:" + strconv.Itoa(port) + entity,
			EntityType: heartbeat.FileType,
			Time:       1585598060,
			UserAgent:  "wakatime/13.0.7",
		},
	})
	require.NoError(t, err)
}

func TestWithDetection_SshConfig_UserKnownHostsFile_Mismatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping because OS is Windows.")
	}

	logs := bytes.NewBuffer(nil)

	teardownLogCapture := captureLogs(logs)
	defer teardownLogCapture()

	shutdown, host, port := testServer(t, true)
	defer shutdown()

	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	ssh_config.DefaultUserSettings = &ssh_config.UserSettings{
		IgnoreErrors: false,
	}

	ssh_config.DefaultUserSettings.ConfigFinder(func() string {
		return tmpFile.Name()
	})

	template, err := os.ReadFile("testdata/ssh_config_userknownhosts")
	require.NoError(t, err)

	knownHostsFile, err := filepath.Abs("./testdata/known_hosts")
	require.NoError(t, err)

	err = os.WriteFile(tmpFile.Name(), []byte(fmt.Sprintf(string(template), host, knownHostsFile)), 0600)
	require.NoError(t, err)

	entity, _ := filepath.Abs("./testdata/main.go")

	sender := mockSender{
		SendHeartbeatsFn: func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			assert.Empty(t, hh)
			return []heartbeat.Result{}, nil
		},
	}

	opts := []heartbeat.HandleOption{
		filter.WithFiltering(filter.Config{
			IncludeOnlyWithProjectFile: true,
		}),
		remote.WithDetection(),
	}

	handle := heartbeat.NewHandle(&sender, opts...)
	results, err := handle([]heartbeat.Heartbeat{
		{
			Category:   heartbeat.CodingCategory,
			Entity:     "ssh://user:pass@github.com:" + strconv.Itoa(port) + entity,
			EntityType: heartbeat.FileType,
			Time:       1585598060,
			UserAgent:  "wakatime/13.0.7",
		},
	})
	require.NoError(t, err)
	assert.Empty(t, results)
	assert.Contains(t, logs.String(), "ssh: handshake failed: ssh: host key mismatch")
}

func TestWithDetection_SshConfig_UserKnownHostsFile_Match(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping because OS is Windows.")
	}

	logs := bytes.NewBuffer(nil)

	teardownLogCapture := captureLogs(logs)
	defer teardownLogCapture()

	shutdown, host, port := testServer(t, true)
	defer shutdown()

	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	ssh_config.DefaultUserSettings = &ssh_config.UserSettings{
		IgnoreErrors: false,
	}

	ssh_config.DefaultUserSettings.ConfigFinder(func() string {
		return tmpFile.Name()
	})

	template, err := os.ReadFile("testdata/ssh_config_userknownhosts")
	require.NoError(t, err)

	knownHostsFile, err := filepath.Abs("./testdata/known_hosts")
	require.NoError(t, err)

	err = os.WriteFile(tmpFile.Name(), []byte(fmt.Sprintf(string(template), host, knownHostsFile)), 0600)
	require.NoError(t, err)

	entity, _ := filepath.Abs("./testdata/main.go")

	sender := mockSender{
		SendHeartbeatsFn: func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			assert.Equal(t, []heartbeat.Heartbeat{
				{
					Category:              heartbeat.CodingCategory,
					Entity:                "ssh://user:pass@example.com:" + strconv.Itoa(port) + entity,
					EntityType:            heartbeat.FileType,
					LocalFile:             hh[0].LocalFile,
					LocalFileNeedsCleanup: true,
					Time:                  1585598060,
					UserAgent:             "wakatime/13.0.7",
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

	opts := []heartbeat.HandleOption{
		filter.WithFiltering(filter.Config{
			Exclude:                    nil,
			Include:                    nil,
			IncludeOnlyWithProjectFile: true,
		}),
		remote.WithDetection(),
	}

	handle := heartbeat.NewHandle(&sender, opts...)
	results, err := handle([]heartbeat.Heartbeat{
		{
			Category:   heartbeat.CodingCategory,
			Entity:     "ssh://user:pass@example.com:" + strconv.Itoa(port) + entity,
			EntityType: heartbeat.FileType,
			Time:       1585598060,
			UserAgent:  "wakatime/13.0.7",
		},
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestWithDetection_Filtered(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping because OS is Windows.")
	}

	logs := bytes.NewBuffer(nil)

	teardownLogCapture := captureLogs(logs)
	defer teardownLogCapture()

	shutdown, host, port := testServer(t, true)
	defer shutdown()

	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	ssh_config.DefaultUserSettings = &ssh_config.UserSettings{
		IgnoreErrors: false,
	}

	ssh_config.DefaultUserSettings.ConfigFinder(func() string {
		return tmpFile.Name()
	})

	template, err := os.ReadFile("testdata/ssh_config_userknownhosts")
	require.NoError(t, err)

	knownHostsFile, err := filepath.Abs("./testdata/known_hosts")
	require.NoError(t, err)

	err = os.WriteFile(tmpFile.Name(), []byte(fmt.Sprintf(string(template), host, knownHostsFile)), 0600)
	require.NoError(t, err)

	entity, _ := filepath.Abs("./testdata/main.go")

	sender := mockSender{
		SendHeartbeatsFn: func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			assert.Empty(t, hh)
			return []heartbeat.Result{}, nil
		},
	}

	opts := []heartbeat.HandleOption{
		filter.WithFiltering(filter.Config{
			Exclude:                    []regex.Regex{regexp.MustCompile(".*")},
			Include:                    nil,
			IncludeOnlyWithProjectFile: true,
		}),
		remote.WithDetection(),
	}

	handle := heartbeat.NewHandle(&sender, opts...)
	results, err := handle([]heartbeat.Heartbeat{
		{
			Category:   heartbeat.CodingCategory,
			Entity:     "ssh://user:pass@example.com:" + strconv.Itoa(port) + entity,
			EntityType: heartbeat.FileType,
			Time:       1585598060,
			UserAgent:  "wakatime/13.0.7",
		},
	})
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestWithCleanup_NotTemporary(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	tmpFile.Close()

	defer os.Remove(tmpFile.Name())

	opts := []heartbeat.HandleOption{
		remote.WithCleanup(),
	}

	sender := mockSender{
		SendHeartbeatsFn: func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			return []heartbeat.Result{
				{
					Status:    201,
					Heartbeat: heartbeat.Heartbeat{},
				},
			}, nil
		},
	}

	handle := heartbeat.NewHandle(&sender, opts...)

	assert.FileExists(t, tmpFile.Name())

	_, err = handle([]heartbeat.Heartbeat{
		{
			LocalFile: tmpFile.Name(),
		},
	})
	require.NoError(t, err)

	assert.FileExists(t, tmpFile.Name())
}

func TestWithCleanup(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	// Should not defer otherwise it will fail on Windows
	tmpFile.Close()

	opts := []heartbeat.HandleOption{
		remote.WithCleanup(),
	}

	sender := mockSender{
		SendHeartbeatsFn: func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			return []heartbeat.Result{
				{
					Status:    201,
					Heartbeat: heartbeat.Heartbeat{},
				},
			}, nil
		},
	}

	handle := heartbeat.NewHandle(&sender, opts...)

	assert.FileExists(t, tmpFile.Name())

	_, err = handle([]heartbeat.Heartbeat{
		{
			LocalFile:             tmpFile.Name(),
			LocalFileNeedsCleanup: true,
		},
	})
	require.NoError(t, err)

	assert.NoFileExists(t, tmpFile.Name())
}

func TestWithCleanup_NotRemoteFile(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer tmpFile.Close()

	opts := []heartbeat.HandleOption{
		remote.WithCleanup(),
	}

	sender := mockSender{
		SendHeartbeatsFn: func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			return []heartbeat.Result{
				{
					Status:    201,
					Heartbeat: heartbeat.Heartbeat{},
				},
			}, nil
		},
	}

	handle := heartbeat.NewHandle(&sender, opts...)

	_, err = handle([]heartbeat.Heartbeat{
		{
			LocalFile: tmpFile.Name(),
		},
	})
	require.NoError(t, err)

	assert.FileExists(t, tmpFile.Name())
}

type mockSender struct {
	SendHeartbeatsFn        func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error)
	SendHeartbeatsFnInvoked bool
}

func (m *mockSender) SendHeartbeats(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
	m.SendHeartbeatsFnInvoked = true
	return m.SendHeartbeatsFn(hh)
}

func keyAuth(_ ssh.ConnMetadata, _ ssh.PublicKey) (*ssh.Permissions, error) {
	permissions := &ssh.Permissions{
		CriticalOptions: map[string]string{},
		Extensions:      map[string]string{},
	}

	return permissions, nil
}

func pwAuth(_ ssh.ConnMetadata, _ []byte) (*ssh.Permissions, error) {
	permissions := &ssh.Permissions{
		CriticalOptions: map[string]string{},
		Extensions:      map[string]string{},
	}

	return permissions, nil
}

func basicServerConfig() *ssh.ServerConfig {
	config := ssh.ServerConfig{
		Config: ssh.Config{
			MACs: []string{"hmac-sha1"},
		},
		PasswordCallback:  pwAuth,
		PublicKeyCallback: keyAuth,
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

	hostPrivateKeySigner, err := ssh.ParsePrivateKey(privKey)
	if err != nil {
		panic(err)
	}

	config.AddHostKey(hostPrivateKeySigner)

	return &config
}

type sshServer struct {
	conn     net.Conn
	config   *ssh.ServerConfig
	sshConn  *ssh.ServerConn
	newChans <-chan ssh.NewChannel
	newReqs  <-chan *ssh.Request
}

func sshServerFromConn(conn net.Conn, config *ssh.ServerConfig) (*sshServer, error) {
	// From a standard TCP connection to an encrypted SSH connection
	sshConn, newChans, newReqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return nil, err
	}

	svr := &sshServer{conn, config, sshConn, newChans, newReqs}
	svr.listenChannels()

	return svr, nil
}

func (svr *sshServer) Wait() error {
	return svr.sshConn.Wait()
}

func (svr *sshServer) Close() error {
	return svr.sshConn.Close()
}

func (svr *sshServer) listenChannels() {
	go func() {
		for chanReq := range svr.newChans {
			go svr.handleChanReq(chanReq)
		}
	}()
	go func() {
		for req := range svr.newReqs {
			go svr.handleReq(req)
		}
	}()
}

func (*sshServer) handleReq(req *ssh.Request) {
	_ = rejectRequest(req)
}

func rejectRequest(req *ssh.Request) error {
	fmt.Printf("ssh rejecting request, type: %s\n", req.Type)

	err := req.Reply(false, []byte{})
	if err != nil {
		fmt.Printf("ssh request reply had error: %v\n", err)
	}

	return err
}

func (chsvr *sshSessionChannelServer) handle() {
	// should maybe do something here...
	go chsvr.handleReqs()
}

func (chsvr *sshSessionChannelServer) handleReqs() {
	for req := range chsvr.newReqs {
		chsvr.handleReq(req)
	}

	fmt.Printf("ssh server session channel complete\n")
}

func (chsvr *sshSessionChannelServer) handleReq(req *ssh.Request) {
	switch req.Type {
	case "env":
		_ = chsvr.handleEnv(req)
	case "subsystem":
		_ = chsvr.handleSubsystem(req)
	default:
		_ = rejectRequest(req)
	}
}

func rejectRequestUnmarshalError(req *ssh.Request, s any, err error) error {
	fmt.Printf("ssh request unmarshaling error, type '%T': %v\n", s, err)

	_ = rejectRequest(req)

	return err
}

type sshEnvRequest struct {
	Envvar string
	Value  string
}

func (chsvr *sshSessionChannelServer) handleEnv(req *ssh.Request) error {
	envReq := &sshEnvRequest{}
	if err := ssh.Unmarshal(req.Payload, envReq); err != nil {
		return rejectRequestUnmarshalError(req, envReq, err)
	}

	_ = req.Reply(true, nil)

	found := false

	for i, envstr := range chsvr.env {
		if strings.HasPrefix(envstr, envReq.Envvar+"=") {
			found = true
			chsvr.env[i] = envReq.Envvar + "=" + envReq.Value
		}
	}

	if !found {
		chsvr.env = append(chsvr.env, envReq.Envvar+"="+envReq.Value)
	}

	return nil
}

func (svr *sshServer) handleChanReq(chanReq ssh.NewChannel) {
	fmt.Printf("channel request: %v, extra: '%v'\n", chanReq.ChannelType(), hex.EncodeToString(chanReq.ExtraData()))

	switch chanReq.ChannelType() {
	case "session":
		if ch, reqs, err := chanReq.Accept(); err != nil {
			fmt.Printf("fail to accept channel request: %v\n", err)

			_ = chanReq.Reject(ssh.ResourceShortage, "channel accept failure")
		} else {
			chsvr := &sshSessionChannelServer{
				sshChannelServer: &sshChannelServer{svr, chanReq, ch, reqs},
				env:              append([]string{}, os.Environ()...),
			}
			chsvr.handle()
		}
	default:
		_ = chanReq.Reject(ssh.UnknownChannelType, "channel type is not a session")
	}
}

type sshSessionChannelServer struct {
	*sshChannelServer
	env []string
}

type sshChannelServer struct {
	svr     *sshServer
	chanReq ssh.NewChannel
	ch      ssh.Channel
	newReqs <-chan *ssh.Request
}

type sshSubsystemRequest struct {
	Name string
}

type sshSubsystemExitStatus struct {
	Status uint32
}

func (chsvr *sshSessionChannelServer) handleSubsystem(req *ssh.Request) error {
	defer func() {
		err1 := chsvr.ch.CloseWrite()
		err2 := chsvr.ch.Close()
		fmt.Printf("ssh server subsystem request complete, err: %v %v\n", err1, err2)
	}()

	subsystemReq := &sshSubsystemRequest{}
	if err := ssh.Unmarshal(req.Payload, subsystemReq); err != nil {
		return rejectRequestUnmarshalError(req, subsystemReq, err)
	}

	// reply to the ssh client

	// no idea if this is actually correct spec-wise.
	// just enough for an sftp server to start.
	if subsystemReq.Name != "sftp" {
		return req.Reply(false, nil)
	}

	_ = req.Reply(true, nil)

	sftpServer, err := sftp.NewServer(chsvr.ch)
	if err != nil {
		return err
	}

	// wait for the session to close
	runErr := sftpServer.Serve()

	exitStatus := uint32(1)

	if runErr == nil {
		exitStatus = uint32(0)
	}

	_, exitStatusErr := chsvr.ch.SendRequest("exit-status", false, ssh.Marshal(sshSubsystemExitStatus{exitStatus}))

	return exitStatusErr
}

func testServer(t *testing.T, expectError bool) (func(), string, int) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	host, portStr, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatal(err)
	}

	shutdown := make(chan struct{})

	go func() {
		for {
			conn, err := listener.Accept()
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

				sshSvr, err := sshServerFromConn(conn, basicServerConfig())
				if err != nil {
					if !expectError {
						t.Error(err)
					}

					return
				}

				_ = sshSvr.Wait()
			}()
		}
	}()

	return func() { close(shutdown); listener.Close() }, host, port
}

func captureLogs(dest io.Writer) func() {
	logOutput := log.Output()

	// will write to log output and dest
	mw := io.MultiWriter(logOutput, dest)

	log.SetOutput(mw)

	return func() {
		log.SetOutput(logOutput)
	}
}
