package log_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestDynamicWriteSyncer_SerializesWrites(t *testing.T) {
	writer := &concurrencyDetectingWriteSyncer{}
	dynamicWriter := log.NewDynamicWriteSyncer(writer)

	const writers = 32

	var wg sync.WaitGroup
	wg.Add(writers)

	for range writers {
		go func() {
			defer wg.Done()

			_, _ = dynamicWriter.Write([]byte("log entry"))
		}()
	}

	wg.Wait()

	assert.False(t, writer.overlapped.Load())
}

type concurrencyDetectingWriteSyncer struct {
	active     atomic.Bool
	overlapped atomic.Bool
}

func (w *concurrencyDetectingWriteSyncer) Write(p []byte) (int, error) {
	if !w.active.CompareAndSwap(false, true) {
		w.overlapped.Store(true)
	}

	time.Sleep(time.Millisecond)
	w.active.Store(false)

	return len(p), nil
}

func (*concurrencyDetectingWriteSyncer) Sync() error {
	return nil
}

var _ zapcore.WriteSyncer = (*concurrencyDetectingWriteSyncer)(nil)
