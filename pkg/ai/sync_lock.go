package ai

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"
	"time"

	"github.com/juju/mutex"
	"github.com/spf13/viper"
	"github.com/wakatime/wakatime-cli/pkg/ini"
)

const (
	syncLockNamePrefix = "wakatime-ai-sync"

	// SyncLockTimeout is intentionally short so editor-triggered heartbeats do
	// not stall when another wakatime-cli process is already syncing AI logs.
	SyncLockTimeout = 100 * time.Millisecond
)

// AcquireSyncLock acquires the process-wide AI sync mutex.
func AcquireSyncLock(ctx context.Context, v *viper.Viper, timeout time.Duration) (mutex.Releaser, error) {
	if v == nil {
		return nil, fmt.Errorf("missing viper instance")
	}

	name, err := syncLockName(ctx, v)
	if err != nil {
		return nil, err
	}

	return mutex.Acquire(mutex.Spec{
		Name:    name,
		Delay:   time.Millisecond,
		Timeout: timeout,
		Cancel:  ctx.Done(),
		Clock:   syncLockClock{delay: time.Millisecond},
	})
}

// IsSyncLockBusy reports whether lock acquisition failed because another
// process is already syncing AI activity.
func IsSyncLockBusy(err error) bool {
	return errors.Is(err, mutex.ErrTimeout) || errors.Is(err, mutex.ErrCancelled)
}

func syncLockName(ctx context.Context, v *viper.Viper) (string, error) {
	internalFilepath, err := ini.InternalFilePath(ctx, v)
	if err != nil {
		return "", fmt.Errorf("failed to load internal config path for ai sync lock: %w", err)
	}

	hash := fnv.New32a()
	_, _ = hash.Write([]byte(internalFilepath))

	return syncLockNamePrefix + "-" + strconv.FormatUint(uint64(hash.Sum32()), 16), nil
}

type syncLockClock struct {
	delay time.Duration
}

func (c syncLockClock) After(time.Duration) <-chan time.Time {
	return time.After(c.delay)
}

func (syncLockClock) Now() time.Time {
	return time.Now()
}
