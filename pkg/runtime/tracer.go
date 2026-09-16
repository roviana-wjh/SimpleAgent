package runtime

import (
	"sync"
)

// Tracer 定义追踪器接口
type Tracer interface {
	Record(trace *Trace) error
	GetTraces() []*Trace
	Reset()
}

// MemoryTracer 基于内存的追踪器
type MemoryTracer struct {
	traces []*Trace
	mu     sync.RWMutex
}

// NewMemoryTracer 创建内存追踪器
func NewMemoryTracer() *MemoryTracer {
	return &MemoryTracer{
		traces: make([]*Trace, 0),
	}
}

// Record 记录追踪
func (t *MemoryTracer) Record(trace *Trace) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.traces = append(t.traces, trace)
	return nil
}

// GetTraces 获取所有追踪
func (t *MemoryTracer) GetTraces() []*Trace {
	t.mu.RLock()
	defer t.mu.RUnlock()

	traces := make([]*Trace, len(t.traces))
	copy(traces, t.traces)
	return traces
}

// Reset 重置追踪
func (t *MemoryTracer) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.traces = make([]*Trace, 0)
}
