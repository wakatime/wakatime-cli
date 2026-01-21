package api

import (
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
)

func TestGroupByAPIURLAndKey(t *testing.T) {
	tests := map[string]struct {
		Input    []heartbeat.Heartbeat
		Expected map[string][]heartbeat.Heartbeat
	}{
		"empty slice": {
			Input:    []heartbeat.Heartbeat{},
			Expected: map[string][]heartbeat.Heartbeat{},
		},
		"single heartbeat": {
			Input: []heartbeat.Heartbeat{
				{
					Entity: "/path/to/file.go",
					APIURL: "https://api.wakatime.com/api/v1",
					APIKey: "00000000-0000-4000-8000-000000000000",
				},
			},
			Expected: map[string][]heartbeat.Heartbeat{
				"https://api.wakatime.com/api/v1|00000000-0000-4000-8000-000000000000": {
					{
						Entity: "/path/to/file.go",
						APIURL: "https://api.wakatime.com/api/v1",
						APIKey: "00000000-0000-4000-8000-000000000000",
					},
				},
			},
		},
		"multiple heartbeats same url and key": {
			Input: []heartbeat.Heartbeat{
				{
					Entity: "/path/to/file1.go",
					APIURL: "https://api.wakatime.com/api/v1",
					APIKey: "00000000-0000-4000-8000-000000000000",
				},
				{
					Entity: "/path/to/file2.go",
					APIURL: "https://api.wakatime.com/api/v1",
					APIKey: "00000000-0000-4000-8000-000000000000",
				},
			},
			Expected: map[string][]heartbeat.Heartbeat{
				"https://api.wakatime.com/api/v1|00000000-0000-4000-8000-000000000000": {
					{
						Entity: "/path/to/file1.go",
						APIURL: "https://api.wakatime.com/api/v1",
						APIKey: "00000000-0000-4000-8000-000000000000",
					},
					{
						Entity: "/path/to/file2.go",
						APIURL: "https://api.wakatime.com/api/v1",
						APIKey: "00000000-0000-4000-8000-000000000000",
					},
				},
			},
		},
		"multiple heartbeats different urls": {
			Input: []heartbeat.Heartbeat{
				{
					Entity: "/path/to/file1.go",
					APIURL: "https://api.wakatime.com/api/v1",
					APIKey: "00000000-0000-4000-8000-000000000000",
				},
				{
					Entity: "/path/to/file2.go",
					APIURL: "https://custom.example.com/api/v1",
					APIKey: "00000000-0000-4000-8000-000000000001",
				},
			},
			Expected: map[string][]heartbeat.Heartbeat{
				"https://api.wakatime.com/api/v1|00000000-0000-4000-8000-000000000000": {
					{
						Entity: "/path/to/file1.go",
						APIURL: "https://api.wakatime.com/api/v1",
						APIKey: "00000000-0000-4000-8000-000000000000",
					},
				},
				"https://custom.example.com/api/v1|00000000-0000-4000-8000-000000000001": {
					{
						Entity: "/path/to/file2.go",
						APIURL: "https://custom.example.com/api/v1",
						APIKey: "00000000-0000-4000-8000-000000000001",
					},
				},
			},
		},
		"same url different keys": {
			Input: []heartbeat.Heartbeat{
				{
					Entity: "/path/to/file1.go",
					APIURL: "https://api.wakatime.com/api/v1",
					APIKey: "00000000-0000-4000-8000-000000000000",
				},
				{
					Entity: "/path/to/file2.go",
					APIURL: "https://api.wakatime.com/api/v1",
					APIKey: "00000000-0000-4000-8000-000000000001",
				},
			},
			Expected: map[string][]heartbeat.Heartbeat{
				"https://api.wakatime.com/api/v1|00000000-0000-4000-8000-000000000000": {
					{
						Entity: "/path/to/file1.go",
						APIURL: "https://api.wakatime.com/api/v1",
						APIKey: "00000000-0000-4000-8000-000000000000",
					},
				},
				"https://api.wakatime.com/api/v1|00000000-0000-4000-8000-000000000001": {
					{
						Entity: "/path/to/file2.go",
						APIURL: "https://api.wakatime.com/api/v1",
						APIKey: "00000000-0000-4000-8000-000000000001",
					},
				},
			},
		},
		"mixed grouping": {
			Input: []heartbeat.Heartbeat{
				{
					Entity: "/work/file1.go",
					APIURL: "https://work.example.com/api/v1",
					APIKey: "00000000-0000-4000-8000-000000000001",
				},
				{
					Entity: "/home/file1.go",
					APIURL: "https://api.wakatime.com/api/v1",
					APIKey: "00000000-0000-4000-8000-000000000000",
				},
				{
					Entity: "/work/file2.go",
					APIURL: "https://work.example.com/api/v1",
					APIKey: "00000000-0000-4000-8000-000000000001",
				},
				{
					Entity: "/home/file2.go",
					APIURL: "https://api.wakatime.com/api/v1",
					APIKey: "00000000-0000-4000-8000-000000000000",
				},
			},
			Expected: map[string][]heartbeat.Heartbeat{
				"https://work.example.com/api/v1|00000000-0000-4000-8000-000000000001": {
					{
						Entity: "/work/file1.go",
						APIURL: "https://work.example.com/api/v1",
						APIKey: "00000000-0000-4000-8000-000000000001",
					},
					{
						Entity: "/work/file2.go",
						APIURL: "https://work.example.com/api/v1",
						APIKey: "00000000-0000-4000-8000-000000000001",
					},
				},
				"https://api.wakatime.com/api/v1|00000000-0000-4000-8000-000000000000": {
					{
						Entity: "/home/file1.go",
						APIURL: "https://api.wakatime.com/api/v1",
						APIKey: "00000000-0000-4000-8000-000000000000",
					},
					{
						Entity: "/home/file2.go",
						APIURL: "https://api.wakatime.com/api/v1",
						APIKey: "00000000-0000-4000-8000-000000000000",
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := groupByAPIURLAndKey(test.Input)

			assert.Equal(t, test.Expected, result)
		})
	}
}

func TestSortKeys(t *testing.T) {
	tests := map[string]struct {
		Input    map[string][]heartbeat.Heartbeat
		Expected []string
	}{
		"empty map": {
			Input:    map[string][]heartbeat.Heartbeat{},
			Expected: []string{},
		},
		"single key": {
			Input: map[string][]heartbeat.Heartbeat{
				"https://api.wakatime.com/api/v1|key1": {},
			},
			Expected: []string{"https://api.wakatime.com/api/v1|key1"},
		},
		"multiple keys sorted": {
			Input: map[string][]heartbeat.Heartbeat{
				"https://z.example.com/api/v1|key1": {},
				"https://a.example.com/api/v1|key2": {},
				"https://m.example.com/api/v1|key3": {},
			},
			Expected: []string{
				"https://a.example.com/api/v1|key2",
				"https://m.example.com/api/v1|key3",
				"https://z.example.com/api/v1|key1",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := sortKeys(test.Input)

			assert.Equal(t, test.Expected, result)
		})
	}
}
