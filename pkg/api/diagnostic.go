package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/wakatime/wakatime-cli/pkg/diagnostic"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/version"
)

type diagnosticsBody struct {
	Platform     string `json:"platform"`
	Architecture string `json:"architecture"`
	Plugin       string `json:"plugin"`
	CliVersion   string `json:"cli_version"`
	Logs         string `json:"logs,omitempty"`
	Stack        string `json:"stacktrace,omitempty"`
}

// SendDiagnostics sends diagnostics to the WakaTime api.
func (c *Client) SendDiagnostics(plugin string, diagnostics ...diagnostic.Diagnostic) error {
	url := c.baseURL + "/plugins/errors"

	log.Debugf("sending diagnostic data to api at %s", url)

	body := diagnosticsBody{
		Platform:     version.OS,
		Architecture: version.Arch,
		CliVersion:   version.Version,
		Plugin:       plugin,
	}

	for _, d := range diagnostics {
		switch d.Type {
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

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return ErrRequest(fmt.Sprintf("failed making request to %q: %s", url, err))
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Err(fmt.Sprintf("failed reading response body from %q: %s", url, err))
	}

	if resp.StatusCode != http.StatusCreated {
		return Err(fmt.Sprintf(
			"invalid response status from %q. got: %d, want: %d. body: %q",
			url,
			resp.StatusCode,
			http.StatusCreated,
			string(respBody),
		))
	}

	return nil
}
