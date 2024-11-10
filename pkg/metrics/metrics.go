package metrics

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// StartProfiling starts profiling cpu and memory. It returns a function that
// should be called to stop profiling and close the files.
func StartProfiling(ctx context.Context) (func(), error) {
	folder, err := ini.WakaResourcesDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed getting user's home directory: %s", err)
	}

	metricsFolder := filepath.Join(folder, "metrics")
	if err := os.MkdirAll(metricsFolder, 0750); err != nil {
		return nil, fmt.Errorf("failed to create metrics folder: %s", err)
	}

	now := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)

	cpuf, err := os.Create(filepath.Join(metricsFolder, fmt.Sprintf("cpu_%s.profile", now))) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to create cpu profile file: %s", err)
	}

	logger := log.Extract(ctx)

	if err := pprof.StartCPUProfile(cpuf); err != nil {
		logger.Errorf("failed to start cpu profile: %s", err)
	}

	memf, err := os.Create(filepath.Join(metricsFolder, fmt.Sprintf("mem_%s.profile", now))) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to create mem profile file: %s", err)
	}

	if err := pprof.WriteHeapProfile(memf); err != nil {
		logger.Errorf("failed to write heap profile: %s", err)
	}

	return func() {
		pprof.StopCPUProfile()

		if err := cpuf.Close(); err != nil {
			logger.Errorf("failed to close cpu profile file: %s", err)
		}

		if err := memf.Close(); err != nil {
			logger.Errorf("failed to close mem profile file: %s", err)
		}
	}, nil
}
