package fileexperts

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/output"
)

type (
	// Data contains the file experts data.
	Data struct {
		Total Total `json:"total"`
		User  User  `json:"user"`
	}

	// FileExperts contains the response for the file_expert endpoint.
	FileExperts struct {
		Data []Data `json:"data"`
	}

	// Entity contains the request for the file-expert endpoint.
	Entity struct {
		Filepath         string  `json:"entity"`
		Project          *string `json:"project"`
		ProjectRootCount *int    `json:"project_root_count"`
	}

	// Total contains the total time spent on a file.
	Total struct {
		Decimal      string  `json:"decimal"`
		Digital      string  `json:"digital"`
		Text         string  `json:"text"`
		TotalSeconds float64 `json:"total_seconds"`
	}

	// User contains the user information.
	User struct {
		ID            string `json:"id"`
		IsCurrentUser bool   `json:"is_current_user"`
		LongName      string `json:"long_name"`
		Name          string `json:"name"`
	}
)

// Caller calls wakatime api to get the file expert.
type Caller interface {
	FileExperts(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error)
}

// NewHandle creates a new Handle, which acts like a processing pipeline,
// with a caller eventually requesting the API.
func NewHandle(caller Caller, opts ...heartbeat.HandleOption) heartbeat.Handle {
	return func(heartbeats []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		var handle heartbeat.Handle = caller.FileExperts
		for i := len(opts) - 1; i >= 0; i-- {
			handle = opts[i](handle)
		}

		return handle(heartbeats)
	}
}

// RenderFileExperts generates a text representation from file experts of the current day.
// If out is set to output.RawJSONOutput or output.JSONOutput, the response will be marshaled to JSON.
// Expects an array of users for the current day. Will return an error otherwise.
func RenderFileExperts(d *FileExperts, out output.Output) (string, error) {
	if d == nil {
		return "", nil
	}

	if len(d.Data) == 0 {
		return "", nil
	}

	if out == output.RawJSONOutput {
		data, err := json.Marshal(d)
		if err != nil {
			return "", fmt.Errorf("failed to marshal json file experts: %s", err)
		}

		return string(data), nil
	}

	var current, other *Data

	for _, data := range d.Data {
		if current != nil && other != nil {
			break
		}

		dev := data

		if data.User.IsCurrentUser {
			current = &dev

			continue
		}

		if other == nil {
			other = &dev
		}
	}

	if out == output.JSONOutput {
		type simplified struct {
			CurrentUser *Data `json:"you,omitempty"`
			OtherUser   *Data `json:"other,omitempty"`
		}

		devs := simplified{
			CurrentUser: current,
			OtherUser:   other,
		}

		data, err := json.Marshal(devs)
		if err != nil {
			return "", fmt.Errorf("failed to marshal json simplified file experts: %s", err)
		}

		return string(data), nil
	}

	var parts []string

	if current != nil {
		parts = append(parts, fmt.Sprintf("You: %s", current.Total.Text))
	}

	if other != nil {
		parts = append(parts, fmt.Sprintf("%s: %s", other.User.Name, other.Total.Text))
	}

	if len(parts) == 0 {
		return "", nil
	}

	return strings.Join(parts, " | "), nil
}
