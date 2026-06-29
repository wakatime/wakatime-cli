package metrics

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

type profilingDeps struct {
	wakaResourcesDir func(context.Context) (string, error)
	mkdirAll         func(string, os.FileMode) error
	createFile       func(string) (io.WriteCloser, error)
	startCPUProfile  func(io.Writer) error
	writeHeapProfile func(io.Writer) error
	stopCPUProfile   func()
}

// StartProfiling starts profiling cpu and memory. It returns a function that
// should be called to stop profiling and close the files.
func StartProfiling(ctx context.Context) (func(), error) {
	return startProfiling(ctx, profilingDeps{
		wakaResourcesDir: ini.WakaResourcesDir,
		mkdirAll:         os.MkdirAll,
		createFile: func(name string) (io.WriteCloser, error) {
			return os.Create(name) //nolint:gosec
		},
		startCPUProfile:  pprof.StartCPUProfile,
		writeHeapProfile: pprof.WriteHeapProfile,
		stopCPUProfile:   pprof.StopCPUProfile,
	})
}

func startProfiling(ctx context.Context, deps profilingDeps) (func(), error) {
	folder, err := deps.wakaResourcesDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed getting user's home directory: %s", err)
	}

	metricsFolder := filepath.Join(folder, "metrics")
	if err := deps.mkdirAll(metricsFolder, 0750); err != nil {
		return nil, fmt.Errorf("failed to create metrics folder: %s", err)
	}

	now := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)

	cpuf, err := deps.createFile(filepath.Join(metricsFolder, fmt.Sprintf("cpu_%s.profile", now)))
	if err != nil {
		return nil, fmt.Errorf("failed to create cpu profile file: %s", err)
	}

	logger := log.Extract(ctx)

	if err := deps.startCPUProfile(cpuf); err != nil {
		logger.Errorf("failed to start cpu profile: %s", err)
	}

	memf, err := deps.createFile(filepath.Join(metricsFolder, fmt.Sprintf("mem_%s.profile", now)))
	if err != nil {
		return nil, fmt.Errorf("failed to create mem profile file: %s", err)
	}

	if err := deps.writeHeapProfile(memf); err != nil {
		logger.Errorf("failed to write heap profile: %s", err)
	}

	return func() {
		deps.stopCPUProfile()

		if err := cpuf.Close(); err != nil {
			logger.Errorf("failed to close cpu profile file: %s", err)
		}

		if err := memf.Close(); err != nil {
			logger.Errorf("failed to close mem profile file: %s", err)
		}
	}, nil
}
