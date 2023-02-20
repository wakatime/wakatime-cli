package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/wakatime/wakatime-cli/pkg/diagnostic"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/version"
)

type diagnosticsBody struct {
	Architecture  string `json:"architecture"`
	CliVersion    string `json:"cli_version"`
	IsPanic       bool   `json:"is_panic,omitempty"`
	Logs          string `json:"logs,omitempty"`
	OriginalError string `json:"error_message,omitempty"`
	Platform      string `json:"platform"`
	Plugin        string `json:"plugin"`
	Stack         string `json:"stacktrace,omitempty"`
}

// SendDiagnostics sends diagnostics to the WakaTime api.
func (c *Client) SendDiagnostics(plugin string, panicked bool, diagnostics ...diagnostic.Diagnostic) error {
	url := c.baseURL + "/plugins/errors"

	log.Debugf("sending diagnostic data to api at %s", url)

	body := diagnosticsBody{
		Architecture: version.Arch,
		CliVersion:   version.Version,
		IsPanic:      panicked,
		Platform:     version.OS,
		Plugin:       plugin,
	}

	for _, d := range diagnostics {
		switch d.Type {
		case diagnostic.TypeError:
			body.OriginalError = d.Value
		case diagnostic.TypeLogs:
			body.Logs = d.Value
		case diagnostic.TypeStack:
			body.Stack = d.Value
		default:
			return fmt.Errorf("unknown diagnostic type %d", d.Type)
		}
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to json marshal request body: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return Err{Err: fmt.Errorf("failed making request to %q: %s", url, err)}
	}
	defer resp.Body.Close() // nolint:errcheck,gosec

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Err{Err: fmt.Errorf("failed reading response body from %q: %s", url, err)}
	}

	if resp.StatusCode != http.StatusCreated {
		return Err{Err: fmt.Errorf(
			"invalid response status from %q. got: %d, want: %d. body: %q",
			url,
			resp.StatusCode,
			http.StatusCreated,
			string(respBody),
		)}
	}

	return nil
}
