package devsforfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/output"
)

type (
	// Devs represents a goal.
	Devs struct {
		Data []Dev `json:"data"`
	}

	// Dev represents the time range of a goal.
	Dev struct {
		User  User  `json:"user"`
		Total Total `json:"total"`
	}

	// Total represents the chart data of a goal.
	Total struct {
		Text string `json:"text"`
	}

	// User represents the owner of a goal.
	User struct {
		Name          string `json:"name"`
		ID            string `json:"id"`
		IsCurrentUser bool   `json:"is_current_user"`
	}

	DevsOutput struct {
		You   Dev `json:"you"`
		Other Dev `json:"other"`
	}
)

// RenderDevsForFile generates a text representation from goal of the current day.
// If out is set to output.RawJSONOutput or output.JSONOutput, the goal will be marshaled to JSON.
// Expects exactly one summary for the current day. Will return an error otherwise.
func RenderDevsForFile(devs *Devs, out output.Output) (string, error) {
	if devs == nil {
		return "", errors.New("no devs found for the current day")
	}

	if out == output.RawJSONOutput {
		data, err := json.Marshal(devs)
		if err != nil {
			return "", fmt.Errorf("failed to marshal json devs: %s", err)
		}

		return string(data), nil
	}

	var currentUser Dev
	for i := range devs.Data {
		if devs.Data[i].User.IsCurrentUser {
			currentUser = devs.Data[i]
			break
		}
	}

	var otherUser Dev
	if len(devs.Data) > 0 {
		otherUser = devs.Data[0]

		if otherUser.User.IsCurrentUser {
			if len(devs.Data) > 1 {
				otherUser = devs.Data[1]
			} else {
				otherUser = Dev{}
			}
		}
	}

	if out == output.JSONOutput {
		output := DevsOutput{}
		if currentUser != (Dev{}) {
			output.You = currentUser
		}
		if otherUser != (Dev{}) {
			output.Other = otherUser
		}

		data, _ := json.Marshal(output)
		return string(data), nil
	}

	var parts []string
	if currentUser != (Dev{}) {
		parts = append(parts, "You: "+currentUser.Total.Text)
	}
	if otherUser != (Dev{}) {
		parts = append(parts, otherUser.User.Name+": "+otherUser.Total.Text)
	}

	return strings.Join(parts, " | "), nil
}
