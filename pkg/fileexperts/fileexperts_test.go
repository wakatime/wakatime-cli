package fileexperts_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/fileexperts"
	"github.com/wakatime/wakatime-cli/pkg/output"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderFileExperts_SingleItem(t *testing.T) {
	d := &fileexperts.FileExperts{
		Data: []fileexperts.Data{
			{
				Total: fileexperts.Total{
					Text: "1 hr 2 mins",
				},
				User: fileexperts.User{
					IsCurrentUser: true,
					LongName:      "John Doe",
					Name:          "John",
				},
			},
		},
	}

	rendered, err := fileexperts.RenderFileExperts(d, output.TextOutput)
	require.NoError(t, err)

	assert.Equal(t, "You: 1 hr 2 mins", rendered)
}

func TestRenderFileExperts_CurrentUser_MultipleItems(t *testing.T) {
	d := &fileexperts.FileExperts{
		Data: []fileexperts.Data{
			{
				Total: fileexperts.Total{
					Text: "2 hrs 16 mins",
				},
				User: fileexperts.User{
					IsCurrentUser: true,
					LongName:      "John Doe",
					Name:          "John",
				},
			},
			{
				Total: fileexperts.Total{
					Text: "1 hr 2 mins",
				},
				User: fileexperts.User{
					LongName: "Karl Marx",
					Name:     "Karl",
				},
			},
			{
				Total: fileexperts.Total{
					Text: "5 mins",
				},
				User: fileexperts.User{
					LongName: "Nick Fury",
					Name:     "Nick",
				},
			},
		},
	}

	rendered, err := fileexperts.RenderFileExperts(d, output.TextOutput)
	require.NoError(t, err)

	assert.Equal(t, "You: 2 hrs 16 mins | Karl: 1 hr 2 mins", rendered)
}

func TestRenderFileExperts_MultipleItems(t *testing.T) {
	d := &fileexperts.FileExperts{
		Data: []fileexperts.Data{
			{
				Total: fileexperts.Total{
					Text: "2 hrs 16 mins",
				},
				User: fileexperts.User{
					LongName: "Karl Marx",
					Name:     "Karl",
				},
			},
			{
				Total: fileexperts.Total{
					Text: "1 hr 2 mins",
				},
				User: fileexperts.User{
					LongName: "Nick Fury",
					Name:     "Nick",
				},
			},
			{
				Total: fileexperts.Total{
					Text: "5 mins",
				},
				User: fileexperts.User{
					IsCurrentUser: true,
					LongName:      "John Doe",
					Name:          "John",
				},
			},
		},
	}

	rendered, err := fileexperts.RenderFileExperts(d, output.TextOutput)
	require.NoError(t, err)

	assert.Equal(t, "You: 5 mins | Karl: 2 hrs 16 mins", rendered)
}

func TestRenderFileExperts_JSON(t *testing.T) {
	data, err := os.ReadFile("testdata/file_experts.json")
	require.NoError(t, err)

	expected, err := os.ReadFile("testdata/file_experts_simplified.json")
	require.NoError(t, err)

	var d fileexperts.FileExperts

	err = json.Unmarshal(data, &d)
	require.NoError(t, err)

	rendered, err := fileexperts.RenderFileExperts(&d, output.JSONOutput)
	require.NoError(t, err)

	assert.Equal(t, string(expected), rendered)
}

func TestRenderFileExperts_RawJSON(t *testing.T) {
	data, err := os.ReadFile("testdata/file_experts.json")
	require.NoError(t, err)

	var d fileexperts.FileExperts

	err = json.Unmarshal(data, &d)
	require.NoError(t, err)

	rendered, err := fileexperts.RenderFileExperts(&d, output.RawJSONOutput)
	require.NoError(t, err)

	assert.Equal(t, string(data), rendered)
}
