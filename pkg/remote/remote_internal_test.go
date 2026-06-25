package remote

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kevinburke/ssh_config"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldDownloadFileFallback(t *testing.T) {
	assert.True(t, shouldDownloadFileFallback(errors.New("connection reset")))
	assert.False(t, shouldDownloadFileFallback(errors.New("ssh: handshake failed: ssh: host key mismatch")))
}

func TestContains(t *testing.T) {
	assert.True(t, contains([]string{"GitHub.com", "example.com"}, "github.COM"))
	assert.False(t, contains([]string{"", "example.com"}, "github.com"))
}

func TestClientUser(t *testing.T) {
	assert.Equal(t, "explicit", Client{User: "explicit"}.user())
	assert.Empty(t, Client{}.user())
}

func TestClientUserFromSSHConfig(t *testing.T) {
	withSSHConfig(t, `
Host original
  User original-user
Host derived
  User derived-user
`)

	assert.Equal(t, "original-user", Client{OriginalHost: "original", Host: "derived"}.user())
	assert.Equal(t, "derived-user", Client{OriginalHost: "alias", Host: "derived"}.user())
}

func TestClientStrictHostKeyChecking(t *testing.T) {
	withSSHConfig(t, `
Host original
  StrictHostKeyChecking accept-new
Host derived
  StrictHostKeyChecking off
`)

	assert.Equal(t, "no", Client{OriginalHost: "original", Host: "derived"}.strictHostKeyChecking())
	assert.Equal(t, "ask", Client{OriginalHost: "alias", Host: "derived"}.strictHostKeyChecking())
}

func TestClientKnownHostsFilesAndKeys(t *testing.T) {
	knownHostsFile, err := filepath.Abs("testdata/known_hosts")
	require.NoError(t, err)

	withSSHConfig(t, "Host example.com\n  UserKnownHostsFile "+knownHostsFile+"\n")

	client := Client{OriginalHost: "example.com", Host: "example.com"}

	assert.Contains(t, client.knownHostsFiles(), knownHostsFile)
	assert.NotEmpty(t, client.knownHostKeys(t.Context()))
}

func TestClientIdentityFile(t *testing.T) {
	identityFile := filepath.Join(t.TempDir(), "id_rsa")
	require.NoError(t, os.WriteFile(identityFile, []byte("not a key"), 0600))

	withSSHConfig(t, `
Host original
  IdentityFile /missing/key
Host derived
  IdentityFile `+identityFile+`
`)

	client := Client{OriginalHost: "original", Host: "derived"}

	assert.Equal(t, identityFile, client.identityFile())
}

func TestClientSignerForIdentityErr(t *testing.T) {
	identityFile := filepath.Join(t.TempDir(), "id_rsa")
	require.NoError(t, os.WriteFile(identityFile, []byte("not a key"), 0600))

	withSSHConfig(t, "Host original\n  IdentityFile "+identityFile+"\n")

	signer, err := Client{OriginalHost: "original", Host: "original"}.signerForIdentity(t.Context())

	require.Error(t, err)
	assert.Nil(t, signer)
	assert.Contains(t, err.Error(), "failed to parse private key")
}

func TestWarnIfUsingRevokedHostKeys(t *testing.T) {
	withSSHConfig(t, "Host original\n  RevokedHostKeys revoked_keys\n")

	var logs bytes.Buffer

	ctx := log.ToContext(context.Background(), log.New(&logs))

	Client{OriginalHost: "original", Host: "original"}.warnIfUsingRevokedHostKeys(ctx)

	assert.Contains(t, logs.String(), "RevokedHostKeys is not supported")
}

func TestHostKeyAlias(t *testing.T) {
	withSSHConfig(t, `
Host original
  HostKeyAlias alias-original
Host derived
  HostKeyAlias alias-derived
`)

	assert.Equal(t, "alias-original", hostKeyAlias(t.Context(), "original", "derived"))
	assert.Equal(t, "alias-derived", hostKeyAlias(t.Context(), "alias", "derived"))
}

func TestWithDetectionSkipsInvalidRemoteAndKeepsLocal(t *testing.T) {
	handle := WithDetection()(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		require.Len(t, hh, 1)
		assert.Equal(t, "/tmp/local.go", hh[0].Entity)

		return []heartbeat.Result{{Status: 201}}, nil
	})

	results, err := handle(t.Context(), []heartbeat.Heartbeat{
		{Entity: "/tmp/local.go", EntityType: heartbeat.FileType},
		{Entity: "ssh://%zz/tmp/main.go", EntityType: heartbeat.FileType},
	})

	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestDeleteLocalFileMissing(t *testing.T) {
	assert.NotPanics(t, func() {
		deleteLocalFile(t.Context(), filepath.Join(t.TempDir(), "missing.go"))
	})
}

func TestNewClientParsesURLAndSSHConfig(t *testing.T) {
	withSSHConfig(t, `
Host original
  HostName derived
  Port 2200
  HostKeyAlias alias
`)

	client, err := NewClient(t.Context(), "sftp://user:pass@original:2222/tmp/main.go")
	require.NoError(t, err)
	assert.Equal(t, "user", client.User)
	assert.Equal(t, "pass", client.Pass)
	assert.Equal(t, "original", client.OriginalHost)
	assert.Equal(t, "derived", client.Host)
	assert.Equal(t, 2222, client.Port)
	assert.Equal(t, "/tmp/main.go", client.Path)
	assert.Equal(t, "alias", client.HostKeyAlias)

	client, err = NewClient(t.Context(), "ssh://original/tmp/main.go")
	require.NoError(t, err)
	assert.Equal(t, 2200, client.Port)

	_, err = NewClient(t.Context(), "ssh://%zz/tmp/main.go")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse remote file url")
}

func TestConnectStrictHostKeyWithoutKnownKey(t *testing.T) {
	withSSHConfig(t, `
Host original
  StrictHostKeyChecking yes
`)

	sshClient, sftpClient, err := Client{OriginalHost: "original", Host: "original", Port: 22}.Connect(t.Context())

	require.Error(t, err)
	assert.Nil(t, sshClient)
	assert.Nil(t, sftpClient)
	assert.Contains(t, err.Error(), "known host key not found")
}

func withSSHConfig(t *testing.T, contents string) {
	t.Helper()

	oldSettings := ssh_config.DefaultUserSettings

	t.Cleanup(func() {
		ssh_config.DefaultUserSettings = oldSettings
	})

	configFile := filepath.Join(t.TempDir(), "ssh_config")
	require.NoError(t, os.WriteFile(configFile, []byte(contents), 0600))

	ssh_config.DefaultUserSettings = &ssh_config.UserSettings{
		IgnoreErrors: false,
	}
	ssh_config.DefaultUserSettings.ConfigFinder(func() string {
		return configFile
	})
}
