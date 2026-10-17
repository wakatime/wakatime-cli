package fileexperts_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/fileexperts"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
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

func TestRenderFileExperts_Empty(t *testing.T) {
	inputs := map[string]*fileexperts.FileExperts{
		"nil response": nil,
		"nil data":     {},
		"empty data":   {Data: []fileexperts.Data{}},
	}
	outputs := map[output.Output]string{
		output.TextOutput:    "",
		output.JSONOutput:    `{}`,
		output.RawJSONOutput: `{"data":[]}`,
	}

	for name, input := range inputs {
		t.Run(name, func(t *testing.T) {
			for out, expected := range outputs {
				t.Run(out.String(), func(t *testing.T) {
					rendered, err := fileexperts.RenderFileExperts(input, out)
					require.NoError(t, err)

					if out == output.TextOutput {
						assert.Equal(t, expected, rendered)
					} else {
						assert.JSONEq(t, expected, rendered)
					}
				})
			}
		})
	}
}

func TestRenderFileExperts_MarshalError(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		for _, out := range []output.Output{output.JSONOutput, output.RawJSONOutput} {
			t.Run(fmt.Sprintf("%s/%v", out, value), func(t *testing.T) {
				d := &fileexperts.FileExperts{Data: []fileexperts.Data{{
					Total: fileexperts.Total{TotalSeconds: value},
				}}}
				rendered, err := fileexperts.RenderFileExperts(d, out)
				require.ErrorContains(t, err, "failed to marshal json")

				expected := `{}`
				if out == output.RawJSONOutput {
					expected = `{"data":[]}`
				}

				assert.JSONEq(t, expected, rendered)
			})
		}
	}
}

func TestRenderFileExperts_TextWithOtherOnlyEmptyUser(t *testing.T) {
	rendered, err := fileexperts.RenderFileExperts(&fileexperts.FileExperts{
		Data: []fileexperts.Data{{}},
	}, output.TextOutput)

	require.NoError(t, err)
	assert.Equal(t, ": ", rendered)
}

func TestNewHandle(t *testing.T) {
	caller := fileExpertsCallerFunc(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, []heartbeat.Heartbeat{{Entity: "entity"}}, hh)

		return []heartbeat.Result{{Status: 201}}, nil
	})

	opt := func(next heartbeat.Handle) heartbeat.Handle {
		return func(ctx context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			hh[0].Entity = "entity"

			return next(ctx, hh)
		}
	}

	results, err := fileexperts.NewHandle(caller, opt)(t.Context(), []heartbeat.Heartbeat{{Entity: "raw"}})
	require.NoError(t, err)
	assert.Equal(t, []heartbeat.Result{{Status: 201}}, results)
}

type fileExpertsCallerFunc func(context.Context, []heartbeat.Heartbeat) ([]heartbeat.Result, error)

func (f fileExpertsCallerFunc) FileExperts(ctx context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
	return f(ctx, hh)
}
