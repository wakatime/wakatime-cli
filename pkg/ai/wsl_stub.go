//go:build !windows

package ai

import "context"

func discoverWSLHomes(context.Context) []string { return nil }
