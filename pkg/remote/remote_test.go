package remote_test

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/remote"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	client, err := remote.NewClient("ssh://wakatime:1234@192.168.1.2:222/home/pi/unicorn-hat/examples/ascii_pic.py")
	require.NoError(t, err)

	assert.Equal(t, remote.Client{
		User: "wakatime",
		Pass: "1234",
		Host: "192.168.1.2",
		Port: 222,
		Path: "/home/pi/unicorn-hat/examples/ascii_pic.py",
	}, client)
}

func TestNewClient_Sftp(t *testing.T) {
	client, err := remote.NewClient("sftp://127.0.0.1")
	require.NoError(t, err)

	assert.Equal(t, remote.Client{
		User: "",
		Pass: "",
		Host: "127.0.0.1",
		Port: 22,
		Path: "",
	}, client)
}

func TestNewClient_Err(t *testing.T) {
	_, err := remote.NewClient("ssh://wakatime:1234@192.168.1.2:port")
	require.Error(t, err)

	assert.EqualError(t, err,
		`failed to parse remote file url: parse "ssh://wakatime:1234@192.168.1.2:port": invalid port ":port" after host`)
}
