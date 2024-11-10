package log

import (
	"sync"

	"go.uber.org/zap/zapcore"
)

// DynamicWriteSyncer allows changing the underlying WriteSyncer at runtime.
type DynamicWriteSyncer struct {
	mu     sync.RWMutex
	writer zapcore.WriteSyncer
}

// Write writes the log entry to the current writer.
func (d *DynamicWriteSyncer) Write(p []byte) (n int, err error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.writer.Write(p)
}

// Sync calls Sync on the current writer.
func (d *DynamicWriteSyncer) Sync() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.writer.Sync()
}

// SetWriter allows updating the underlying writer at runtime.
func (d *DynamicWriteSyncer) SetWriter(newWriter zapcore.WriteSyncer) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.writer = newWriter
}

// NewDynamicWriteSyncer initializes the dynamic writer with an initial writer.
func NewDynamicWriteSyncer(initial zapcore.WriteSyncer) *DynamicWriteSyncer {
	return &DynamicWriteSyncer{writer: initial}
}
