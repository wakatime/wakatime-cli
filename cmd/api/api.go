package api

import (
	"fmt"
	"strings"

	paramscmd "github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/log"

	tz "github.com/gandarez/go-olson-timezone"
)

// NewClient initializes a new api client with all options following the
// passed in parameters.
func NewClient(params paramscmd.API) (*api.Client, error) {
	withAuth, err := api.WithAuth(api.BasicAuth{
		Secret: params.Key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set up auth option on api client: %w", err)
	}

	return newClient(params, withAuth)
}

// NewClientWithoutAuth initializes a new api client with all options following the
// passed in parameters and disabled authentication.
func NewClientWithoutAuth(params paramscmd.API) (*api.Client, error) {
	return newClient(params)
}

// newClient contains the logic of client initialization, except auth initialization.
func newClient(params paramscmd.API, opts ...api.Option) (*api.Client, error) {
	opts = append(opts, api.WithTimeout(params.Timeout))
	opts = append(opts, api.WithHostname(params.Hostname))

	tz, err := tz.Name()
	if err != nil {
		log.Debugf("failed to detect local timezone: %s", err)
	} else {
		opts = append(opts, api.WithTimezone(tz))
	}

	if params.DisableSSLVerify {
		opts = append(opts, api.WithDisableSSLVerify())
	}

	if !params.DisableSSLVerify && params.SSLCertFilepath != "" {
		withSSLCert, err := api.WithSSLCertFile(params.SSLCertFilepath)
		if err != nil {
			return nil, fmt.Errorf("failed to set up ssl cert file option on api client: %s", err)
		}

		opts = append(opts, withSSLCert)
	} else if !params.DisableSSLVerify {
		withSSLCert, err := api.WithSSLCertPool(api.CACerts())
		if err != nil {
			return nil, fmt.Errorf("failed to set up ssl cert pool option on api client: %s", err)
		}

		opts = append(opts, withSSLCert)
	}

	if params.ProxyURL != "" {
		withProxy, err := api.WithProxy(params.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("failed to set up proxy option on api client: %w", err)
		}

		opts = append(opts, withProxy)

		if strings.Contains(params.ProxyURL, `\\`) {
			withNTLMRetry, err := api.WithNTLMRequestRetry(params.ProxyURL)
			if err != nil {
				return nil, fmt.Errorf("failed to set up ntlm request retry option on api client: %w", err)
			}

			opts = append(opts, withNTLMRetry)
		}
	}

	opts = append(opts, api.WithUserAgent(params.Plugin))

	return api.NewClient(params.URL, opts...), nil
}
