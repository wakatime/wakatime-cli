//go:build ip

package main_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	Meta struct {
		Data MetaData `json:"data"`
	}

	MetaData struct {
		Ips MetaDataIps `json:"ips"`
	}

	MetaDataIps struct {
		Api MetaDataIpsApi `json:"api"`
	}

	MetaDataIpsApi struct {
		V4 []string `json:"v4"`
		V6 []string `json:"v6"`
	}
)

func TestHardCodedIPs(t *testing.T) {
	resp, err := http.Get("https://wakatime.com/api/v1/meta")
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var meta Meta

	err = json.Unmarshal(body, &meta)
	require.NoError(t, err)

	assert.Contains(t, meta.Data.Ips.Api.V4, api.BaseIPAddrv4)
	assert.Contains(t, meta.Data.Ips.Api.V6, api.BaseIPAddrv6)
}
