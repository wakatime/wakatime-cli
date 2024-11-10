package api

import (
	"context"
	"fmt"
	"strings"

	paramscmd "github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/log"

	tz "github.com/gandarez/go-olson-timezone"
)

// NewClient initializes a new api client with all options following the
// passed in parameters.
func NewClient(ctx context.Context, params paramscmd.API) (*api.Client, error) {
	withAuth, err := api.WithAuth(api.BasicAuth{
		Secret: params.Key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set up auth option on api client: %w", err)
	}

	return newClient(ctx, params, withAuth)
}

// NewClientWithoutAuth initializes a new api client with all options following the
// passed in parameters and disabled authentication.
func NewClientWithoutAuth(ctx context.Context, params paramscmd.API) (*api.Client, error) {
	return newClient(ctx, params)
}

// newClient contains the logic of client initialization, except auth initialization.
func newClient(ctx context.Context, params paramscmd.API, opts ...api.Option) (*api.Client, error) {
	opts = append(opts, api.WithTimeout(params.Timeout))
	opts = append(opts, api.WithHostname(strings.TrimSpace(params.Hostname)))

	logger := log.Extract(ctx)

	tz, err := timezone()
	if err != nil {
		logger.Debugf("failed to detect local timezone: %s", err)
	} else {
		opts = append(opts, api.WithTimezone(strings.TrimSpace(tz)))
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
		opts = append(opts, api.WithSSLCertPool(api.CACerts(ctx)))
	}

	if params.ProxyURL != "" {
		withProxy, err := api.WithProxy(params.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("failed to set up proxy option on api client: %w", err)
		}

		opts = append(opts, withProxy)

		if strings.Contains(params.ProxyURL, `\\`) {
			withNTLMRetry, err := api.WithNTLMRequestRetry(ctx, params.ProxyURL)
			if err != nil {
				return nil, fmt.Errorf("failed to set up ntlm request retry option on api client: %w", err)
			}

			opts = append(opts, withNTLMRetry)
		}
	}

	opts = append(opts, api.WithUserAgent(ctx, params.Plugin))

	return api.NewClient(params.URL, opts...), nil
}

func timezone() (name string, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panicked when detecting timezone: %s", e)
		}
	}()

	name, err = tz.Name()

	return name, err
}
